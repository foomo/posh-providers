package browser

import (
	"context"
	"os"
	"os/exec"
	"runtime"

	"github.com/foomo/posh-providers/pkg/proxy"
	"github.com/foomo/posh/pkg/command/tree"
	env "github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Command struct {
	l           log.Logger
	cfg         Config
	proxyCfg    proxy.Config
	commandTree tree.Root
}

func NewCommand(l log.Logger) (*Command, error) {
	inst := &Command{
		l: l.Named("browser"),
	}

	if err := viper.UnmarshalKey("browser", &inst.cfg); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("proxies", &inst.proxyCfg); err != nil {
		return nil, err
	}

	if inst.cfg.ConfigPath != "" {
		if err := files.MkdirAll(inst.cfg.ConfigPath); err != nil {
			return nil, errors.Wrapf(err, "failed to create config path: %s", inst.cfg.ConfigPath)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "browser",
		Description: "Open a browser window through a proxy tunnel",
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

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	envName := r.Args().At(0)
	url := r.Args().At(1)

	// If the first arg is not a known env name, treat it as a URL.
	if envName != "" {
		if _, err := c.cfg.Environment(envName); err != nil {
			url = envName
			envName = ""
		}
	}

	// Resolve: env → cfg
	var proxyName string

	incognito := false

	if e, err := c.cfg.Environment(envName); err == nil {
		proxyName = e.Proxy

		if url == "" {
			url = e.URL
		}

		incognito = e.Incognito
	}

	if proxyName == "" {
		proxyName = c.cfg.Proxy
	}

	if !incognito {
		incognito = c.cfg.Incognito
	}

	if proxyName == "" {
		return errors.New("no proxy configured; set browser.proxy or browser.environments.<name>.proxy in config")
	}

	addr, sshPID, stop, err := c.proxyCfg.Addr(ctx, c.l, proxyName)
	if err != nil {
		return err
	}

	dataDir, err := os.MkdirTemp(env.Path(c.cfg.ConfigPath), "posh-browser-*")
	if err != nil {
		stop()

		return err
	}

	binary := c.cfg.Command
	if binary == "" {
		binary = defaultBrowserCommand()
	}

	args := []string{
		"--proxy-server=socks5://" + addr,
		"--user-data-dir=" + dataDir,
		"--no-first-run",
		"--no-default-browser-check",
	}
	if incognito {
		args = append(args, "--incognito")
	}

	if url != "" {
		args = append(args, url)
	}

	cmd := exec.CommandContext(ctx, binary, args...)

	if err := cmd.Start(); err != nil {
		stop()

		_ = os.RemoveAll(dataDir)

		return err
	}

	// Watcher: survives posh being killed. Polls Chrome's PID and kills the SSH
	// tunnel when Chrome exits, ensuring no orphaned tunnel processes.
	watcher := newWatcherCmd(cmd.Process.Pid, sshPID)
	_ = watcher.Start()

	go func() {
		_ = cmd.Wait()

		if watcher.Process != nil {
			_ = watcher.Process.Kill()
		}

		stop()

		_ = os.RemoveAll(dataDir)
	}()

	return nil
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
