package gcloud

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/command/tree"
	env2 "github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

type (
	Command struct {
		l             log.Logger
		name          string
		op            *onepassword.OnePassword
		gcloud        *GCloud
		kubectl       *kubectl.Kubectl
		commandTree   *tree.Root
		clusterNameFn ClusterNameFn
	}
	ClusterNameFn func(environment Environment, cluster Cluster) string
	CommandOption func(command *Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

func CommandWithOnePassword(v *onepassword.OnePassword) CommandOption {
	return func(o *Command) {
		o.op = v
	}
}

func CommandWithClusterNameFn(v ClusterNameFn) CommandOption {
	return func(o *Command) {
		o.clusterNameFn = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, gcloud *GCloud, kubectl *kubectl.Kubectl, opts ...CommandOption) *Command {
	inst := &Command{
		l:       l.Named("gcloud"),
		name:    "gcloud",
		gcloud:  gcloud,
		kubectl: kubectl,
		clusterNameFn: func(environment Environment, cluster Cluster) string {
			ret := environment.Name
			if cluster.Name != DefaultCluster {
				ret = ret + "-" + cluster.Name
			}
			if cluster.DefaultRole() != DefaultRole {
				ret = cluster.DefaultRole() + "@" + ret
			}
			return ret
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}
	inst.commandTree = &tree.Root{
		Name:        inst.name,
		Description: "Run google cloud sdk commands",
		Node: &tree.Node{
			Execute: inst.execute,
		},
		Nodes: tree.Nodes{
			{
				Name:        "login",
				Description: "Login to gcloud",
				Execute:     inst.authLogin,
			},
			{
				Name:        "docker",
				Description: "Configure docker access",
				Execute:     inst.authConfigureDocker,
			},
			{
				Name:        "kubeconfig",
				Description: "Retrieve kube config",
				Nodes: tree.Nodes{
					{
						Name:        "environment",
						Description: "Name of the environment",
						Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.gcloud.cfg.EnvironmentNames())
						},
						Nodes: tree.Nodes{
							{
								Name:        "cluster",
								Description: "Name of the cluster",
								Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
									account, err := inst.gcloud.cfg.Environment(r.Args().At(1))
									if err != nil {
										return nil
									}
									return suggests.List(account.ClusterNames())
								},
								Execute: inst.containerClustersGetCredentials,
							},
						},
					},
				},
			},
		},
	}

	return inst
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.commandTree.Name
}

func (c *Command) Description() string {
	return c.commandTree.Description
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help() string {
	return `Execute google-cloud-sdk commands.

Usage:
  gcloud [cmd]

Available commands:
	login                        Login into your google cloud account
  docker                       Configure docker access
  kubeconfig [env] [cluster]   Retrieve kube config for the given cluster
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "gcloud").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.PassThroughFlags()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) authLogin(ctx context.Context, r *readline.Readline) error {
	if err := shell.New(ctx, c.l, "gcloud", "auth", "login").
		Args(r.AdditionalArgs()...).
		Run(); err != nil {
		return err
	}
	return nil
}

func (c *Command) authConfigureDocker(ctx context.Context, r *readline.Readline) error {
	if err := shell.New(ctx, c.l, "gcloud", "auth", "configure-docker").
		Args(r.AdditionalArgs()...).
		Run(); err != nil {
		return err
	}
	return nil
}

func (c *Command) containerClustersGetCredentials(ctx context.Context, r *readline.Readline) error {
	var env []string
	environment, err := c.gcloud.cfg.Environment(r.Args().At(1))
	if err != nil {
		return errors.Errorf("failed to retrieve environment for: %s", r.Args().At(1))
	}

	cluster, err := environment.Cluster(r.Args().At(2))
	if err != nil {
		return errors.Errorf("failed to retrieve cluster for: %s", r.Args().At(2))
	}

	// resolve or retrieve service account access token
	accessTokenFilename := path.Join(
		os.Getenv(env2.ProjectRoot),
		c.gcloud.cfg.AccessTokenPath,
		fmt.Sprintf("%s@%s-%s.json", cluster.DefaultRole(), environment.Name, cluster.Name),
	)
	if stat, err := os.Stat(accessTokenFilename); err == nil && !stat.IsDir() {
		c.l.Debug("using existing access token file:", accessTokenFilename)
		env = c.gcloud.EnvWithAccessToken(env, accessTokenFilename)
	} else if cluster.AccessToken != nil {
		if c.op == nil {
			return errors.New("missing OnePassword provider to retrieve configured access token")
		}
		// retrieve token and write to file
		if value, err := c.op.GetDocument(ctx, *cluster.AccessToken); err != nil {
			return errors.Wrap(err, "failed to retrieve access token")
		} else if err := files.MkdirAll(c.gcloud.cfg.AccessTokenPath); err != nil {
			return errors.Wrap(err, "failed to create access token path")
		} else if err := os.WriteFile(accessTokenFilename, []byte(value), 0600); err != nil {
			return errors.Wrap(err, "failed to write access token")
		}
		c.l.Debug("retrieved and store access token file:", accessTokenFilename)
		env = c.gcloud.EnvWithAccessToken(env, accessTokenFilename)
	}

	kubectlCluster := c.kubectl.Cluster(c.clusterNameFn(environment, cluster))

	return shell.New(ctx, c.l, "gcloud", "container", "clusters", "get-credentials",
		"--project", environment.Project,
		"--region", cluster.Region,
		cluster.DefaultFullName(),
	).
		Args(r.AdditionalArgs()...).
		Env(kubectlCluster.Env()).
		Env(env...).
		Run()
}
