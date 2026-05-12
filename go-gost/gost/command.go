package gost

import (
	"context"
	"os/exec"

	gokaziconfig "github.com/foomo/gokazi/pkg/config"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
		gk          *gokazi.Gokazi
		cfg         Config
		name        string
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

func NewCommand(l log.Logger, gk *gokazi.Gokazi, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("gost"),
		gk:        gk,
		name:      "gost",
		configKey: "gost",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	for name, value := range inst.cfg {
		inst.gk.Add("gost."+name, gokaziconfig.Task{
			Name:        "gost",
			Description: value,
			Path:        "",
			Cwd:         env.ProjectRoot(),
			Args: []string{
				"-C", value,
			},
		})
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Manage gost processes",
		Nodes: tree.Nodes{
			{
				Name:        "start",
				Description: "Start a gost process",
				Args: tree.Args{
					{
						Name:        "name",
						Description: "Config names",
						Repeat:      true,
						Optional:    true,
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.cfg.Names())
						},
					},
				},
				Execute: inst.start,
			},
			{
				Name:        "stop",
				Description: "Stop a gost process",
				Args: tree.Args{
					{
						Name:        "name",
						Description: "Config names",
						Repeat:      true,
						Optional:    true,
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.cfg.Names())
						},
					},
				},
				Execute: inst.stop,
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

func (c *Command) start(ctx context.Context, r *readline.Readline) error {
	names := c.cfg.Names()
	if r.Args().LenGt(2) {
		names = r.Args().From(1)
	}

	c.l.Info("starting gost processes:")

	for _, name := range names {
		path, err := c.cfg.Path(name)
		if err != nil {
			return err
		}

		c.l.Info("└  " + name + ": " + path)

		if err := c.gk.Start(ctx, "gost."+name, exec.CommandContext(ctx, "gost", "-C", path)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) stop(ctx context.Context, r *readline.Readline) error {
	names := c.cfg.Names()
	if r.Args().LenGt(2) {
		names = r.Args().From(1)
	}

	c.l.Info("stopping gost processes:")

	for _, name := range names {
		if _, err := c.cfg.Path(name); err != nil {
			return err
		}

		c.l.Info("└  " + name)

		if err := c.gk.Stop(ctx, "gost."+name); err != nil {
			return err
		}
	}

	return nil
}
