package gcloud

import (
	"context"

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
				Name:        "login",
				Description: "Login into your google cloud account",
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
  kubeconfig <cluster>   Retrieve kube config
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

func (c *Command) completeClusters(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
	var ret []goprompt.Suggest
	for _, cluster := range c.gcloud.cfg.ClusterNames() {
		ret = append(ret, goprompt.Suggest{Text: cluster})
	}
	return ret
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
	var clusters []Cluster
	if r.Args().Len() == 0 {
		clusters = c.gcloud.cfg.Clusters
	} else {
		for _, arg := range r.Args() {
			if value, ok := c.gcloud.cfg.ClustersByName(arg); ok {
				clusters = append(clusters, value)
			}
		}
	}
	for _, value := range clusters {
		cluster := c.kubectl.Cluster(value.Name)
		if err := shell.New(ctx, c.l, "gcloud", "container", "clusters", "get-credentials",
			c.gcloud.cfg.Project,
			"--region", value.Region,
		).
			Args(r.AdditionalArgs()...).
			Env(cluster.Env()).
			Run(); err != nil {
			return err
		}
	}
	return nil
}
