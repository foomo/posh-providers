package gnat

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/foomo/posh-providers/kubernetes/kubectl"
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

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

type (
	// Command is the posh command that drives the gnat terminal UI against a
	// configured NATS JetStream server profile.
	Command struct {
		l           log.Logger
		cfg         Config
		name        string
		configKey   string
		kubectl     *kubectl.Kubectl
		commandTree tree.Root
	}
	// CommandOption configures a Command in NewCommand.
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// WithName sets the name the command is registered under, overriding the
// default "gnat".
func WithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

// WithConfigKey sets the viper key the command reads its Config from,
// overriding the default "gnat".
func WithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// CommandWithKubectl sets the kubectl client used to open a port-forward for
// profiles that configure urlForward. Required only for those profiles.
func CommandWithKubectl(kubectl *kubectl.Kubectl) CommandOption {
	return func(o *Command) {
		o.kubectl = kubectl
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// NewCommand creates a Command, unmarshalling its Config from viper under
// configKey ("gnat" unless overridden by WithConfigKey). It returns an error
// if the config cannot be unmarshalled.
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

// Name returns the name the command is registered under.
func (c *Command) Name() string {
	return c.commandTree.Node().Name
}

// Description returns the command's one-line description.
func (c *Command) Description() string {
	return c.commandTree.Node().Description
}

// Complete returns shell autocompletion suggestions for the current input.
func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

// Execute runs the gnat terminal UI for the profile named in r's arguments.
func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

// Help returns the command's help text for the current input.
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
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names NATS and JetStream concepts rather than the tool, since a caller asks
// about streams and consumers without knowing which client is wired up.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when inspecting or administering a NATS JetStream server - browsing " +
			"streams, consumers, subjects or messages, purging a stream, or checking what a " +
			"configured NATS profile points at. Note the tool is an interactive terminal UI, so " +
			"it cannot be driven unattended.",
	}
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

	url := profile.URL

	if c.kubectl != nil && profile.URLForward != nil {
		cluster := c.kubectl.Cluster(name)
		if cluster == nil {
			return errors.New("no cluster found")
		}

		var cancel context.CancelFunc

		ctx, url, cancel = cluster.PortForward(ctx, *profile.URLForward)
		defer cancel()
	}

	args := []string{"-url", url}
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
