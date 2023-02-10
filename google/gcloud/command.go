package gcloud

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
)

type Command struct {
	l           log.Logger
	gcloud      *GCloud
	kubectl     *kubectl.Kubectl
	commandTree *tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, gcloud *GCloud, kubectl *kubectl.Kubectl) *Command {
	inst := &Command{
		l:       l.Named("gcloud"),
		gcloud:  gcloud,
		kubectl: kubectl,
	}

	inst.commandTree = &tree.Root{
		Name:        "gcloud",
		Description: "Run google cloud sdk commands",
		Node: &tree.Node{
			Execute: inst.execute,
		},
		Nodes: tree.Nodes{
			{
				Name:        "environment",
				Description: "Environments to access",
				Values:      inst.completeAccounts,
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
						Args: tree.Args{
							{
								Name:     "cluster",
								Repeat:   true,
								Optional: true,
								Suggest:  inst.completeClusters,
							},
						},
						Execute: inst.containerClustersGetCredentials,
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
	login                  Login into your google cloud account
  docker                 Configure docker access
  kubeconfig <cluster>    Retrieve kube config
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

// Auto-complete clusters based on selected account
func (c *Command) completeClusters(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
	var ret []goprompt.Suggest

	account := r.Args().At(0)
	for _, acc := range c.gcloud.cfg.Environments {
		if strings.Contains(account, acc.Name) {
			for _, cluster := range acc.Clusters {
				ret = append(ret, goprompt.Suggest{Text: cluster.Name})
			}
		}
	}

	return ret
}

func (c *Command) completeAccounts(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	accounts, err := c.gcloud.ParseAccounts(ctx)
	if err != nil {
		c.l.Debug("failed to walk files", err.Error())
		return nil
	}

	var suggestions []goprompt.Suggest
	for _, acc := range accounts {
		suggestions = append(suggestions, goprompt.Suggest{
			Text:        acc.Environment,
			Description: fmt.Sprintf("%q cluster with role %q", acc.Cluster, acc.Role),
		})
	}
	return suggestions
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
	environment := r.Args().At(0)
	clusterName := r.Args().At(2)

	clusters := c.gcloud.cfg.ClusterNamesForEnv(environment)
	if clusterName != "" {
		clusters = []string{clusterName}
	}

	for _, cluster := range clusters {
		serviceAccounts, err := c.gcloud.FindAccounts(ctx, environment, cluster)
		if err != nil {
			return err
		}

		if len(serviceAccounts) > 1 {
			c.l.Warnf("multiple accounts found for env %q and cluster %q", environment, cluster)
		}
		accountPath, _ := filepath.Abs(serviceAccounts[0].Path)

		kubectlCluster := c.kubectl.Cluster(environment + "-" + cluster)
		gcloudCluster, ok := c.gcloud.cfg.FindCluster(environment, cluster)
		if !ok {
			return fmt.Errorf("could not find configuration for env %q and cluster %q", environment, cluster)
		}

		sh := shell.New(ctx, c.l, "gcloud", "container", "clusters", "get-credentials",
			"--project", gcloudCluster.Project,
			"--region", gcloudCluster.Region,
			cluster,
		).
			Args(r.AdditionalArgs()...).
			Env("GOOGLE_APPLICATION_CREDENTIALS=" + accountPath).
			Env("CLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE=" + accountPath).
			Env("GOOGLE_CREDENTIALS=" + accountPath).
			Env(kubectlCluster.Env())

		if err := sh.Run(); err != nil {
			return err
		}
	}
	return nil
}
