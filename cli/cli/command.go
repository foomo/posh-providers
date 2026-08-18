package cli

import (
	"context"
	_ "embed"

	"github.com/foomo/go/options"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/exec"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
)

//go:embed SKILL.md
var skill string

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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog shows two
// sibling subcommands; what it cannot show is that one is read-only while the
// other is an interactive browser flow that hangs an agent, nor that
// authentication here is a process-wide GITHUB_TOKEN other commands depend on.
func (c *Command) Skill(ctx context.Context) string {
	return skill
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
