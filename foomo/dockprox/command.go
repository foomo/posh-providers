package dockprox

import (
	"context"
	_ "embed"
	"os/exec"

	gokaziconfig "github.com/foomo/gokazi/pkg/config"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
)

//go:embed SKILL.md
var skill string

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
		Args:        []string{inst.cfg.Config},
	})

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Manage the dockprox background process",
		Nodes: tree.Nodes{
			{
				Name:        "start",
				Description: "Start dockprox as a background process using the configured config file",
				Execute:     inst.start,
			},
			{
				Name:        "stop",
				Description: "Stop the running dockprox process; does not find one started by menubar",
				Execute:     inst.stop,
			},
			{
				Name:        "menubar",
				Description: "Start the dockprox menubar app; needs a desktop session and cannot be stopped by stop",
				Execute:     inst.menubar,
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

// Skill implements the optional command.Skiller interface. The rendered tree
// cannot show that this manages a background process outliving the command, that
// `stop` cannot find one started by `menubar` because the registered task
// matches on the config-path argument only `start` passes, or that what the
// proxy binds lives in the config file rather than anywhere in the tree.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) start(ctx context.Context, r *readline.Readline) error {
	c.l.Info("starting dockprox")

	return c.gk.Start(ctx, "dockprox", exec.CommandContext(ctx, "dockprox", "serve", "--config", c.cfg.Config))
}

func (c *Command) menubar(ctx context.Context, r *readline.Readline) error {
	c.l.Info("starting dockprox menubar")

	return c.gk.Start(ctx, "dockprox", exec.CommandContext(ctx, "dockprox", "menubar"))
}

func (c *Command) stop(ctx context.Context, r *readline.Readline) error {
	return c.gk.Stop(ctx, "dockprox")
}
