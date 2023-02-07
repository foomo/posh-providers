package gcloud

import (
	"context"
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
				Name:        "accounts",
				Description: "List of accounts to access",
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
	accounts, err := ParseAccounts(ctx, c.gcloud.cfg.ConfigDir)
	if err != nil {
		c.l.Debug("failed to walk files", err.Error())
		return nil
	}

	var suggestions []goprompt.Suggest
	for _, acc := range accounts {
		suggestions = append(suggestions, goprompt.Suggest{
			Text:        acc.Environment,
			Description: acc.Path,
		})
	}
	return suggestions
}

func (c *Command) toAccountLabel(accountPath string) string {
	return strings.TrimSuffix(strings.TrimSpace(filepath.Base(accountPath)), ".json")
}

func (c *Command) fromAccountLabel(label string) string {
	return filepath.Join(c.gcloud.cfg.ConfigDir, label+".json")
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
	accountName := r.Args().At(0)
	clusterName := r.Args().At(2)

	if clusterName == "" {
		// All Clusters
	}

	kubectlCluster := c.kubectl.Cluster(accountName)
	gcloudCluster := c.gcloud.cfg.Environments[0].Clusters[0]

	if err := shell.New(ctx, c.l, "gcloud", "container", "clusters", "get-credentials",
		"--project", gcloudCluster.Project,
		"--region", gcloudCluster.Region,
		clusterName,
	).
		Args(r.AdditionalArgs()...).
		Env(kubectlCluster.Env()).
		Env("GOOGLE_APPLICATION_CREDENTIALS=" + c.fromAccountLabel(accountName)).
		Run(); err != nil {
		return err
	}
	return nil
}
