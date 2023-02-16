package k9s

import (
	"context"
	"fmt"

	"github.com/foomo/posh-providers/foomo/squadron"
	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
)

type (
	Command struct {
		l           log.Logger
		kubectl     *kubectl.Kubectl
		squadron    *squadron.Squadron
		commandTree *tree.Root
		namespaceFn NamespaceFn
	}
	NamespaceFn   func(cluster, fleet, squadron string) string
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithNamespaceFn(v NamespaceFn) CommandOption {
	return func(o *Command) {
		o.namespaceFn = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, kubectl *kubectl.Kubectl, squadron *squadron.Squadron, opts ...CommandOption) *Command {
	inst := &Command{
		l:        l.Named("k9s"),
		kubectl:  kubectl,
		squadron: squadron,
		namespaceFn: func(cluster, fleet, squadron string) string {
			if fleet == "default" {
				return squadron
			} else {
				return fmt.Sprintf("%s-%s", fleet, squadron)
			}
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}
	inst.commandTree = &tree.Root{
		Name:        "k9s",
		Description: "open the k9s dashboard",
		Node: &tree.Node{
			Args: tree.Args{
				{
					Name:    "cluster",
					Suggest: inst.completeClusters,
				},
				{
					Name:    "fleet",
					Suggest: inst.completeFleets,
				},
				{
					Name:    "squadron",
					Suggest: inst.completeSquadrons,
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

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return `Open the k9s dashboard.

Usage:
  k9s [cluster] [namespace]

Examples:
  k9s example-cluster my-namespace
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	cluster, fleet, squad := r.Args().At(0), r.Args().At(1), r.Args().At(2)
	return shell.New(ctx, c.l, "k9s", "-n", c.namespaceFn(cluster, fleet, squad), "--logoless").
		Args(r.AdditionalArgs()...).
		Env(c.kubectl.Cluster(cluster).Env()).
		Run()
}

func (c *Command) completeClusters(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
	return suggests.List(c.kubectl.Clusters())
}

func (c *Command) completeFleets(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
	if cluster, ok := c.squadron.Cluster(r.Args().At(0)); ok {
		return suggests.List(cluster.Fleets)
	}
	return nil
}

func (c *Command) completeSquadrons(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
	if value, err := c.squadron.List(); err != nil {
		c.l.Debug(err.Error())
		return nil
	} else {
		return suggests.List(value)
	}
}
