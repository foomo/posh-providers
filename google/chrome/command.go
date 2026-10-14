package chrome

import (
	"context"
	_ "embed"
	"os/exec"
	"runtime"
	"strings"

	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
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
	// Command is the posh command that launches Chrome against a configured
	// profile, each with its own isolated user-data directory.
	Command struct {
		l           log.Logger
		cfg         Config
		name        string
		configKey   string
		commandTree tree.Root
	}
	// CommandOption configures a Command in NewCommand.
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// CommandWithName sets the name the command is registered under, overriding
// the default "chrome".
func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

// CommandWithConfigKey sets the viper key the command reads its Config from,
// overriding the default "chrome".
func CommandWithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// NewCommand creates a Command, unmarshalling its Config from viper under
// configKey ("chrome" unless overridden by WithConfigKey) and creating the
// configured profile directory if needed. It returns an error if the config
// cannot be unmarshalled or the directory cannot be created.
func NewCommand(l log.Logger, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("chrome"),
		name:      "chrome",
		configKey: "chrome",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	if inst.cfg.Path != "" {
		if err := files.MkdirAll(inst.cfg.Path); err != nil {
			return nil, errors.Wrapf(err, "failed to create config path: %s", inst.cfg.Path)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Open Chrome with an isolated profile directory",
		Args: tree.Args{
			{
				Name:        "profile",
				Description: "Profile name from the chrome config; also names its data directory",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.cfg.ProfileNames())
				},
			},
			{
				Name:        "url",
				Description: "URL to open, overriding the profile's configured url",
				Optional:    true,
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

// Execute launches Chrome for the profile and optional URL named in r's
// arguments.
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
// cannot show that each profile keeps a persistent logged-in session in its own
// user-data dir, that the command returns before the browser has really
// started, or that a profile's proxy is passed verbatim rather than resolved.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// leads with needing a graphical session, since the most useful thing an agent
// can learn from this file is that it cannot drive the command at all.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when a task needs a browser opened against one of this project's " +
			"environments - a staging or production URL, an admin UI, a session behind a " +
			"proxy or tunnel, or a profile whose stored logins and cookies must persist. " +
			"Requires a graphical session, so it cannot be run unattended.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	profile, err := c.cfg.Profile(r.Args().At(0))
	if err != nil {
		return err
	}

	url := r.Args().At(1)
	if url == "" {
		url = profile.URL
	}

	binary := c.cfg.Command
	if binary == "" {
		binary = defaultBrowserCommand()
	}

	args := []string{
		"--user-data-dir=" + env.Path(c.cfg.Path, r.Args().At(0)),
		"--no-default-browser-check",
		"--no-first-run",
	}

	if c.cfg.Incognito || profile.Incognito {
		args = append(args, "--incognito")
	}

	if profile.Proxy != "" {
		args = append(args, "--proxy-server="+profile.Proxy)
	}

	if url != "" {
		if profile.App {
			args = append(args, "--app="+url)
		} else {
			args = append(args, url)
		}
	}

	return exec.CommandContext(ctx, binary, args...).Start()
}

func defaultBrowserCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	case "windows":
		return `C:\Program Files\Google\Chrome\Application\chrome.exe`
	default:
		return "google-chrome"
	}
}
