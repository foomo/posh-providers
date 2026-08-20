package gnat

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/spf13/viper"
)

//go:embed SKILL.md
var skill string

type (
	Command struct {
		l           log.Logger
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

func WithName(v string) CommandOption {
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

func NewCommand(l log.Logger, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("gnat"),
		name:      "gnat",
		configKey: "gnat",
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
		Description: "Open the gnat terminal UI against a configured NATS JetStream server",
		Args: tree.Args{
			{
				Name:        "name",
				Description: "Profile name from the profiles config, resolved to its url",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					var ret []goprompt.Suggest

					for name, profile := range inst.cfg.Profiles {
						ret = append(ret, goprompt.Suggest{Text: name, Description: profile.Description})
					}

					return ret
				},
			},
		},
		Execute: inst.execute,
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
// cannot show that this is a blocking full-screen TUI unusable unattended, that
// the profile silently selects which NATS server is mutable from inside it, or
// that `configDir` scopes gnat's own config files to the project rather than the
// user's home directory.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	name := r.Args().At(0)

	profile, err := c.cfg.Profile(name)
	if err != nil {
		return err
	}

	args := []string{"-url", profile.URL}
	if theme := c.cfg.Theme; theme != "" {
		args = append(args, "-theme", theme)
	}

	var envs []string
	if dir := c.cfg.ConfigDir; dir != "" {
		envs = append(envs, fmt.Sprintf("XDG_CONFIG_HOME=%s", dir))
	}

	return shell.New(ctx, c.l, "gnat").
		Args(args...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Env(envs...).
		Run()
}
