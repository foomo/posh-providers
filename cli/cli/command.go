package cli

import (
	"context"

	"github.com/foomo/go/options"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/exec"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
)

type Command struct {
	l           log.Logger
	cli         *CLI
	name        string
	commandTree tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) options.Option[*Command] {
	return func(o *Command) {
		o.name = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cli *CLI, opts ...options.Option[*Command]) *Command {
	inst := &Command{
		l:    l.Named("gh"),
		cli:  cli,
		name: "gh",
	}

	options.Apply(inst, opts...)

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run gh",
		Nodes: tree.Nodes{
			{
				Name:        "auth",
				Description: "Authenticate gh and git with GitHub",
				Nodes: tree.Nodes{
					{
						Name:        "status",
						Description: "View authentication status",
						Execute:     inst.execAuth,
					},
					{
						Name:        "refresh",
						Description: "Refresh stored authentication credentials",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("reset-token", false, "Reset the stored auth token")
							fs.Default().StringArray("scopes", nil, "Additional authentication scopes (repeatable)")

							return nil
						},
						Execute: inst.execAuth,
					},
				},
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

func (c *Command) execAuth(ctx context.Context, r *readline.Readline) error {
	return exec.NewCommand(ctx, "gh").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}
