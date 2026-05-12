package dockprox

import (
	"context"
	"os/exec"

	gokaziconfig "github.com/foomo/gokazi/pkg/config"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
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
		l:         l.Named("dockprox"),
		gk:        gk,
		name:      "dockprox",
		configKey: "dockprox",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	inst.gk.Add("dockprox", gokaziconfig.Task{
		Name:        "dockprox",
		Description: inst.cfg.Config,
		Args:        []string{"serve"},
	})

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Manage dockprox processes",
		Nodes: tree.Nodes{
			{
				Name:        "start",
				Description: "Start a dockprox process",
				Execute:     inst.start,
			},
			{
				Name:        "stop",
				Description: "Stop a dockprox process",
				Execute:     inst.stop,
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
	c.l.Info("starting dockprox process:")

	return c.gk.Start(ctx, "dockprox", exec.CommandContext(ctx, "dockprox", "serve", "--config", c.cfg.Config))
}

func (c *Command) stop(ctx context.Context, r *readline.Readline) error {
	return c.gk.Stop(ctx, "dockprox")
}
