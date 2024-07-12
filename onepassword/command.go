package onepassword

import (
	"context"
	"os"
	"path"

	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
)

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
		Description: "Execute 1Password commands",
		Execute:     inst.auth,
		Nodes: tree.Nodes{
			{
				Name:        "auth",
				Description: "Sign into your account",
				Execute:     inst.auth,
			},
			{
				Name:        "get",
				Description: "Retrieve an item",
				Args: tree.Args{
					{
						Name:        "id",
						Description: "Item name or uuid",
					},
				},
				Execute: inst.get,
			},
			{
				Name:        "download",
				Description: "Download a document",
				Args: tree.Args{
					{
						Name:        "id",
						Description: "Item name or uuid",
					},
					{
						Name:        "output",
						Description: "Save the document to the file path instead of stdout",
					},
				},
				Execute: inst.download,
			},
			{
				Name:        "register",
				Description: "Register an account",
				Args: tree.Args{
					{
						Name:        "email",
						Description: "User email address",
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
