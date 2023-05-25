package cdktf

import (
	"context"
	"errors"
	"os/exec"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
		cfg         Config
		name        string
		cache       cache.Namespace
		configKey   string
		commandTree tree.Root
	}
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

func WithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:     l.Named("cdktf"),
		name:  "cdktf",
		cache: cache.Get("cdktf"),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}
	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run cdktf",
		Nodes: tree.Nodes{
			{
				Name:        "list",
				Description: "List stacks in app",
				Execute:     inst.list,
			},
			{
				Name:        "diff",
				Description: "Perform a diff (terraform plan) for the given stack",
				Execute:     inst.diff,
			},
			{
				Name:        "deploy",
				Description: "Deploy the given stacks",
				Execute:     inst.deploy,
			},
			{
				Name:        "destroy",
				Description: "Destroy the given stacks",
				Execute:     inst.destroy,
			},
			{
				Name:        "output",
				Description: "Prints the output of stacks",
				Execute:     inst.output,
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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	if _, err := exec.LookPath("cdktf"); err != nil {
		c.l.Print()
		return errors.New(`
Please ensure you have the cdktf installed!

- Install binary:

$ npm install --global cdktf-cli
        `)
	}
	return nil
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) list(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "cdktf", "list").
		Dir(c.cfg.Path).
		Run()
}

func (c *Command) diff(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "cdktf", "diff").
		Dir(c.cfg.Path).
		Run()
}

func (c *Command) deploy(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "cdktf", "deploy").
		Dir(c.cfg.Path).
		Run()
}

func (c *Command) destroy(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "cdktf", "destroy").
		Dir(c.cfg.Path).
		Run()
}

func (c *Command) output(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "cdktf", "output").
		Dir(c.cfg.Path).
		Run()
}

//nolint:forcetypeassert
//func (c *Command) paths(ctx context.Context) []string {
//	return c.cache.Get("paths", func() any {
//		if value, err := files.Find(ctx, ".", "cdktf.yml"); err != nil {
//			c.l.Debug("failed to walk files", err.Error())
//			return []string{}
//		} else {
//			return value
//		}
//	}).([]string)
//}
