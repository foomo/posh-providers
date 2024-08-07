package gcloud

import (
	"context"
	"fmt"
	"os"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
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
		commandTree   tree.Root
		clusterNameFn ClusterNameFn
	}
	ClusterNameFn func(name string, cluster Cluster) string
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
		clusterNameFn: func(name string, cluster Cluster) string {
			return name
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}
	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run google cloud sdk commands",
		Execute:     inst.execute,
		Nodes: tree.Nodes{
			{
				Name:        "login",
				Description: "Login to gcloud",
				Args: tree.Args{
					{
						Name:     "account",
						Optional: true,
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.gcloud.cfg.AccountNames())
						},
					},
				},
				Execute: inst.authLogin,
			},
			{
				Name:        "docker",
				Description: "Configure docker access",
				Execute:     inst.authConfigureDocker,
			},
			{
				Name:        "kubeconfig",
				Description: "Retrieve kube config for the given cluster",
				Args: tree.Args{
					{
						Name: "cluster",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.gcloud.cfg.ClusterNames())
						},
					},
				},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("profile", "", "Store credentials in given profile.")
					return fs.Internal().SetValues("profile", "gcloud")
				},
				Execute: inst.containerClustersGetCredentials,
			},
		},
	})

	return inst
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.commandTree.Node().Name
}

func (c *Command) Description() string {
	return c.commandTree.Node().Description
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "gcloud").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) authLogin(ctx context.Context, r *readline.Readline) error {
	accountName := r.Args().At(1)
	account, err := c.gcloud.cfg.Account(accountName)
	if err != nil {
		return err
	}

	// resolve or retrieve service account access token
	keyFilename := env.Path(
		c.gcloud.ServiceAccountKeysPath(),
		fmt.Sprintf("%s.json", accountName),
	)
	if err := files.Exists(keyFilename); err == nil {
		c.l.Debug("using existing access token file:", keyFilename)
		return c.authLoginServiceAccount(ctx, r, keyFilename)
	}

	if account.Key != nil && c.op != nil { // retrieve token and write to file
		if value, err := c.op.GetDocument(ctx, *account.Key); err != nil {
			return errors.Wrap(err, "failed to retrieve service account key")
		} else if err := files.MkdirAll(c.gcloud.ServiceAccountKeysPath()); err != nil {
			return errors.Wrap(err, "failed to create service account key path")
		} else if err := os.WriteFile(keyFilename, []byte(value), 0600); err != nil {
			return errors.Wrap(err, "failed to write service account key")
		}
		c.l.Debug("retrieved and store service account key file:", keyFilename)
		return c.authLoginServiceAccount(ctx, r, keyFilename)
	} else {
		c.l.Debug("using default login")
		return c.authLoginDefault(ctx, r)
	}
}

func (c *Command) authLoginDefault(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "gcloud", "auth", "login").
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) authLoginServiceAccount(ctx context.Context, r *readline.Readline, keyFilename string) error {
	return shell.New(ctx, c.l, "gcloud",
		"auth", "activate-service-account",
		"--key-file", keyFilename,
	).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) authConfigureDocker(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "gcloud", "auth", "configure-docker").
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) containerClustersGetCredentials(ctx context.Context, r *readline.Readline) error {
	var args []string
	ifs := r.FlagSets().Internal()
	clusterName := r.Args().At(1)

	cluster, err := c.gcloud.cfg.Cluster(clusterName)
	if err != nil {
		return errors.Errorf("failed to retrieve cluster for: %q", clusterName)
	}

	kubectlCluster := c.kubectl.Cluster(c.clusterNameFn(clusterName, cluster))
	if kubectlCluster == nil {
		return errors.Errorf("failed to retrieve kubectl cluster for: %q", cluster.Name)
	}

	if cluster.Account != "" {
		if account, err := c.gcloud.cfg.Account(cluster.Account); err != nil {
			return errors.Errorf("failed to retrieve account for: %q", cluster.Account)
		} else {
			args = append(args, "--account", account.Name)
		}
	}

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	return shell.New(ctx, c.l, "gcloud", "container", "clusters", "get-credentials", cluster.Name,
		"--project", cluster.Project,
		"--region", cluster.Region,
	).
		Args(args...).
		Args(r.AdditionalArgs()...).
		Env(kubectlCluster.Env(profile)).
		Run()
}
