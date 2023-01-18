package onepassword

import (
	"context"

	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
)

type (
	Command struct {
		l           log.Logger
		op          *OnePassword
		commandTree *tree.Root
	}
	CommandOption func(*Command) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, op *OnePassword, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:  l.Named("onePassword"),
		op: op,
	}
	for _, opt := range opts {
		if opt != nil {
			if err := opt(inst); err != nil {
				return nil, err
			}
		}
	}
	inst.commandTree = &tree.Root{
		Name: "op",
		Nodes: []*tree.Node{
			{
				Name:        "get",
				Description: "retrieve item",
				Args: []*tree.Arg{
					{
						Name: "id",
					},
				},
				Execute: inst.get,
			},
			{
				Name:        "signin",
				Description: "sign into your account",
				Execute:     inst.signin,
			},
			{
				Name:        "register",
				Description: "register an account",
				Args: []*tree.Arg{
					{
						Name: "email",
					},
				},
				Execute: inst.register,
			},
		},
	}
	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.commandTree.Name
}

func (c *Command) Description() string {
	return "run go mod"
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline, d prompt.Document) []prompt.Suggest {
	return c.commandTree.RunCompletion(ctx, r)
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.RunExecution(ctx, r)
}

func (c *Command) Help() string {
	return `1Password session helper.

Usage:
  op [command]

Available commands:
  get [id]          Retrieve an entry from your account
  signin            Sign into your 1Password account for the session
  register [email]  Add your 1Password account
`
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

func (c *Command) register(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l,
		"op", "account", "add",
		"--address", c.op.cfg.Account+".1password.eu",
		"--email", r.Args().At(1),
	).
		Args(r.AdditionalArgs()...).
		Wait()
}

func (c *Command) signin(ctx context.Context, r *readline.Readline) error {
	if ok, _ := c.op.Session(); ok {
		c.l.Info("Already signed in")
		return nil
	} else if err := c.op.SignIn(ctx); err != nil {
		return err
	}
	return nil
}
