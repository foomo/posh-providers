package kubeprompt

import (
	"context"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
)

type Command struct {
	l           log.Logger
	kubectl     *kubectl.Kubectl
	commandTree *tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, kubectl *kubectl.Kubectl) *Command {
	inst := &Command{
		l:       l.Named("kubeprompt"),
		kubectl: kubectl,
	}
	inst.commandTree = &tree.Root{
		Name:        "kubeprompt",
		Description: "open the kubectl prompt",
		Node: &tree.Node{
			Args: tree.Args{
				{
					Name:    "cluster",
					Suggest: inst.completeClusters,
				},
			},
			Execute: inst.execute,
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
	return `Open interactive kubectl prompt.

Usage:
  kubeprompt [cluster]

Examples:
  kubeprompt example-cluster
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "kube-prompt").
		Args(r.AdditionalArgs()...).
		Env(c.kubectl.Cluster(r.Args().At(0)).Env()).
		Run()
}

func (c *Command) completeClusters(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
	return suggests.List(c.kubectl.Clusters())
}
