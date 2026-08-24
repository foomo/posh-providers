package dockprox

import (
	"context"
	_ "embed"
	"errors"
	"os/exec"
	"strings"

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

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

// Task ids registered with gokazi. They are distinct because the two verbs run
// different command lines and gokazi identifies a process by its args.
const (
	taskServe   = "dockprox.serve"
	taskMenubar = "dockprox.menubar"
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

	// One task per verb: gokazi matches a running process by requiring every
	// registered arg to appear in its command line, and `menubar` runs a bare
	// `dockprox menubar` with no config path. Registering a single task with the
	// config arg meant a menubar process was never found - so `stop` reported
	// nothing running and the already-running guard let `start` launch a second
	// dockprox on the same ports.
	inst.gk.Add(taskServe, gokaziconfig.Task{
		Name:        "dockprox",
		Description: inst.cfg.Config,
		Args:        []string{inst.cfg.Config},
	})
	inst.gk.Add(taskMenubar, gokaziconfig.Task{
		Name:        "dockprox",
		Description: "menubar",
		Args:        []string{"menubar"},
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
				Description: "Stop the running dockprox process, whether started by start or menubar",
				Execute:     inst.stop,
			},
			{
				Name:        "menubar",
				Description: "Start the dockprox menubar app; needs a desktop session",
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
// `menubar` needs a graphical session, or that what the proxy binds lives in the
// config file rather than anywhere in the tree.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The triggers name
// the symptom - a container hostname or proxied port not resolving - because the
// proxy is infrastructure nobody thinks about until it is down.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when this project's local Docker reverse proxy needs starting or " +
			"stopping, or when a container hostname or proxied local port is not " +
			"resolving, refusing connections, or reporting an address already in use. " +
			"Also for the dockprox menubar app.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) start(ctx context.Context, r *readline.Readline) error {
	c.l.Info("starting dockprox")

	return c.gk.Start(ctx, taskServe, exec.CommandContext(ctx, "dockprox", "serve", "--config", c.cfg.Config))
}

func (c *Command) menubar(ctx context.Context, r *readline.Readline) error {
	c.l.Info("starting dockprox menubar")

	return c.gk.Start(ctx, taskMenubar, exec.CommandContext(ctx, "dockprox", "menubar"))
}

// stop stops whichever dockprox is running: `start` and `menubar` register
// different tasks, so both have to be tried. Both ids are always registered, so
// the one that is not running reports ErrNotRunning - expected here, not a
// failure. Only a genuine error aborts.
func (c *Command) stop(ctx context.Context, r *readline.Readline) error {
	var stopped bool

	for _, id := range []string{taskServe, taskMenubar} {
		if err := c.gk.Stop(ctx, id); err != nil {
			if errors.Is(err, gokazi.ErrNotRunning) {
				continue
			}

			return err
		}

		stopped = true
	}

	if !stopped {
		c.l.Info("no dockprox process running")
	}

	return nil
}
