package chrome

import (
	"context"
	"os/exec"
	"runtime"

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

func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

func CommandWithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

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

	if inst.cfg.ConfigPath != "" {
		if err := files.MkdirAll(inst.cfg.ConfigPath); err != nil {
			return nil, errors.Wrapf(err, "failed to create config path: %s", inst.cfg.ConfigPath)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Open a Google Chrome browser",
		Args: tree.Args{
			{
				Name:     "env",
				Optional: true,
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.cfg.EnvironmentNames())
				},
			},
			{
				Name:     "url",
				Optional: true,
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

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	environ, err := c.cfg.Environment(r.Args().At(0))
	if err != nil {
		return err
	}

	url := r.Args().At(1)
	if url == "" {
		url = environ.URL
	}

	binary := c.cfg.Command
	if binary == "" {
		binary = defaultBrowserCommand()
	}

	args := []string{

		"--user-data-dir=" + env.Path(c.cfg.ConfigPath, r.Args().At(0)),
		"--no-default-browser-check",
		"--no-first-run",
	}

	if c.cfg.Incognito || environ.Incognito {
		args = append(args, "--incognito")
	}

	if environ.Proxy != "" {
		args = append(args, "--proxy-server="+environ.Proxy)
	}

	if url != "" {
		args = append(args, url)
	}

	return exec.CommandContext(ctx, binary, args...).Run()
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
