package onepassword

import (
	"context"
	_ "embed"
	"os"
	"path"

	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
)

//go:embed SKILL.md
var skill string

type (
	Command struct {
		l           log.Logger
		op          *OnePassword
		name        string
		commandTree tree.Root
	}
	CommandOption func(*Command) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) CommandOption {
	return func(o *Command) error {
		o.name = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, op *OnePassword, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:    l.Named("onePassword"),
		op:   op,
		name: "op",
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(inst); err != nil {
				return nil, err
			}
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Sign in to the configured 1Password account for this shell session",
		Execute:     inst.auth,
		Nodes: tree.Nodes{
			{
				Name:        "auth",
				Description: "Sign in interactively and cache the session for this shell",
				Execute:     inst.auth,
			},
			{
				Name:        "get",
				Description: "Print an item as JSON, including its secret fields",
				Args: tree.Args{
					{
						Name:        "id",
						Description: "Item name or UUID within the configured account",
					},
				},
				Execute: inst.get,
			},
			{
				Name:        "download",
				Description: "Save a document item to a file, creating parent directories",
				Args: tree.Args{
					{
						Name:        "id",
						Description: "Item name or UUID within the configured account",
					},
					{
						Name:        "output",
						Description: "File path to write the document to; parent directories are created",
					},
				},
				Execute: inst.download,
			},
			{
				Name:        "register",
				Description: "Add the configured account to the local op CLI, interactively",
				Args: tree.Args{
					{
						Name:        "email",
						Description: "Email address of the 1Password account to add",
					},
				},
				Execute: inst.register,
			},
		},
	})

	return inst, nil
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

// Skill implements the optional command.Skiller interface. The catalog cannot
// show that `get` prints secret field values, that `auth` is interactive and
// echoes the session token, that a bare `op` signs in rather than printing help,
// or that several other providers resolve their credentials through this one and
// fail here when the session lapses.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) get(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l,
		"op",
		"--account", c.op.cfg.Account,
		"item", "get", r.Args().At(1),
		"--format", "json",
	).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) download(ctx context.Context, r *readline.Readline) error {
	if err := os.MkdirAll(path.Dir(r.Args().At(2)), 0700); err != nil {
		return err
	}

	return shell.New(ctx, c.l,
		"op",
		"--account", c.op.cfg.Account,
		"document", "get", r.Args().At(1),
		"--out-file", r.Args().At(2),
	).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) register(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l,
		"op", "account", "add",
		"--address", c.op.cfg.Account+".1password.eu",
		"--email", r.Args().At(1),
	).
		Args(r.AdditionalArgs()...).
		Wait()
}

func (c *Command) auth(ctx context.Context, r *readline.Readline) error {
	if ok, _ := c.op.IsAuthenticated(ctx); ok {
		c.l.Info("Already signed in")
		return nil
	} else if err := c.op.SignIn(ctx); err != nil {
		return err
	}

	return nil
}
