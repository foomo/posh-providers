package webdriverio

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
		op          *onepassword.OnePassword
		cfg         Config
		name        string
		cache       cache.Namespace
		configKey   string
		commandTree *tree.Root
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

func NewCommand(l log.Logger, c cache.Cache, op *onepassword.OnePassword, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("wdio"),
		name:      "wdio",
		cache:     c.Get("wdio"),
		configKey: "webdriverio",
		op:        op,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}
	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	inst.commandTree = &tree.Root{
		Name:        inst.name,
		Description: "run wdio commands",
		Nodes: tree.Nodes{
			{
				Name:        "mode",
				Description: "run mode",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.cfg.Modes.Keys())
				},
				Nodes: tree.Nodes{
					{
						Name:        "site",
						Description: "configured site",
						Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.cfg.Sites.Keys())
						},
						Nodes: tree.Nodes{
							{
								Name:        "env",
								Description: "configured env",
								Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
									if value, ok := inst.cfg.Sites[r.Args().At(1)]; ok {
										return suggests.List(value.Keys())
									}
									return nil
								},
								Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
									fs.String("tag", "", "run suite on specific tag")
									fs.String("spec", "", "run suite on specific specs")
									fs.String("suite", "", "run suite on test suite")
									fs.String("scenario", "", "run suite on specific specs")
									fs.String("log-level", "info", "set the log level")
									fs.Bool("headless", false, "run suite in headless mode")
									fs.Bool("debug", false, "run in debug mode and leave browser open after test failure")
									fs.Bool("bail", false, "stop test runner after specific amount of tests have failed")
									if r.Args().LenGte(4) {
										if err := fs.SetValues("spec", inst.specs(ctx, r.Args().At(3))...); err != nil {
											return err
										}
										if err := fs.SetValues("tag", inst.tags(ctx, r.Args().At(3), fs.GetString("spec"))...); err != nil {
											return err
										}
										if err := fs.SetValues("scenario", inst.scenarios(ctx, r.Args().At(3), fs.GetString("spec"))...); err != nil {
											return err
										}
									}
									if err := fs.SetValues("log-level", "info", "warn", "debug"); err != nil {
										return err
									}
									return nil
								},
								Args: tree.Args{
									{
										Name:     "path",
										Repeat:   false,
										Optional: true,
										Suggest: func(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
											return suggests.List(inst.paths(ctx))
										},
									},
								},
								Execute: inst.execute,
							},
						},
					},
				},
			},
		},
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.commandTree.Name
}

func (c *Command) Description() string {
	return c.commandTree.Description
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return `Run wdio commands.

Usage:
  wdio [mode] [site] [env] [path]
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	mode, site, env := r.Args().At(0), r.Args().At(1), r.Args().At(2)
	siteConfig := c.cfg.Sites[site][env]
	modeConfig := c.cfg.Modes[mode]

	envs := []string{
		"LOG_LEVEL=" + r.FlagSet().GetString("log-level"),
		"NODE_TLS_REJECT_UNAUTHORIZED=0", // allow TLS errors when in local mode with self-signed certificates
	}
	var args []string

	if r.FlagSet().GetBool("debug") {
		envs = append(envs, fmt.Sprintf("debug=%s", "true"))
	}
	if r.FlagSet().GetBool("headless") {
		envs = append(envs, fmt.Sprintf("HEADLESS=%s", "true"))
	}
	if r.FlagSet().GetBool("ci") {
		envs = append(envs, fmt.Sprintf("E2E_ENV=%s", "ci"))
	} else {
		envs = append(envs, fmt.Sprintf("E2E_ENV=%s", "chromium"))
	}
	if value := r.FlagSet().GetString("scenario"); value != "" {
		envs = append(envs, fmt.Sprintf("SCENARIOS=%s", strings.Trim(value, "\"")))
	}
	if value := r.FlagSet().GetString("tag"); value != "" {
		args = append(args, "--cucumberOpts.tagExpression", "'"+strings.Trim(value, "\"")+"'")
	}
	if value := r.FlagSet().GetString("spec"); value != "" {
		args = append(args, "--spec", value)
	}
	if value := r.FlagSet().GetString("suite"); value != "" {
		args = append(args, "--suite", value)
	}
	// base url
	baseURL := siteConfig.Domain
	if modeConfig.HostPrefix != "" {
		baseURL = modeConfig.HostPrefix + "." + baseURL
	}
	if modeConfig.Port != "" {
		baseURL += ":" + modeConfig.Port
	}
	envs = append(envs, fmt.Sprintf("E2E_BASE_URL=%s", baseURL))
	// basic auth
	if siteConfig.AuthUsername != nil && siteConfig.AuthPassword != nil {
		if username, err := c.op.Get(ctx, *siteConfig.AuthUsername); err != nil {
			return err
		} else if password, err := c.op.Get(ctx, *siteConfig.AuthPassword); err != nil {
			return err
		} else {
			envs = append(envs, fmt.Sprintf("BASIC_AUTH=%s:%s", url.QueryEscape(username), url.QueryEscape(password)))
		}
	}

	var dirs []string
	if r.Args().LenGte(4) {
		dirs = append(dirs, r.Args().At(3))
	} else {
		dirs = c.paths(ctx)
	}

	c.l.Info("Running wdio...")
	for _, dir := range dirs {
		c.l.Info("└  " + dir)
		if err := shell.New(ctx, c.l, "wdio", "run", "e2e/wdio.conf.ts").
			Args(args...).
			Args(r.PassThroughFlags()...).
			Args(r.AdditionalArgs()...).
			Dir(dir).
			Env(envs...).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		values, err := files.Find(ctx, ".", "wdio.conf.ts", files.FindWithIgnore(`^\.`, "node_modules"))
		if err != nil {
			return []string{}
		}
		ret := make([]string, 0, len(values))
		for _, file := range values {
			dir := strings.TrimSuffix(file, "/e2e/wdio.conf.ts")
			if _, err := os.Stat(dir); err != nil {
				continue
			}
			ret = append(ret, dir)
		}
		return ret
	}).([]string)
}

//nolint:forcetypeassert
func (c *Command) specs(ctx context.Context, dir string) []string {
	return c.cache.Get("spec-"+strings.ReplaceAll(dir, "/", "-"), func() any {
		if value, err := files.Find(ctx, dir+"/e2e/features", "*.feature"); err != nil {
			return []string{}
		} else if value == nil {
			return []string{}
		} else {
			for i, s := range value {
				value[i] = strings.TrimPrefix(s, dir+"/e2e/features/")
			}
			return value
		}
	}).([]string)
}

//nolint:forcetypeassert
func (c *Command) tags(ctx context.Context, dir, spec string) []string {
	cacheKey := strings.ReplaceAll(dir, "/", "-")
	filename := "'*.feature'"
	if spec != "" {
		filename = spec
		cacheKey += "-" + strings.TrimSuffix(spec, ".feature")
	}
	return c.cache.Get("tag-"+cacheKey, func() any {
		res, err := shell.New(ctx, c.l, "find", dir+"/e2e/features", "-type", "f", "-name", filename, "-exec", "cat", "{}", "\\;", "|", "grep", "'^\\t@'").CombinedOutput()
		if err != nil {
			return []string{}
		}
		resMap := map[string]bool{}
		for _, value := range strings.Split(string(bytes.TrimSpace(res)), "\n") {
			for _, tag := range strings.Split(strings.TrimSpace(value), " ") {
				if _, ok := resMap[tag]; !ok {
					resMap[tag] = true
				}
			}
		}
		ret := make([]string, 0, len(resMap))
		for tag := range resMap {
			ret = append(ret, tag)
		}
		return ret
	}).([]string)
}

//nolint:forcetypeassert
func (c *Command) scenarios(ctx context.Context, dir, spec string) []string {
	cacheKey := strings.ReplaceAll(dir, "/", "-")
	filename := "'*.feature'"
	if spec != "" {
		filename = spec
		cacheKey += "-" + strings.TrimSuffix(spec, ".feature")
	}
	return c.cache.Get("scenario-"+cacheKey, func() any {
		dir += "/e2e/features"
		res, err := shell.New(ctx, c.l, "find", dir, "-type", "f", "-name", filename, "-exec", "cat", "{}", "\\;", "|", "grep", "'^\\tScenario'").CombinedOutput()
		if err != nil {
			return []string{}
		}
		resMap := map[string]bool{}
		for _, value := range strings.Split(string(bytes.TrimSpace(res)), "\n") {
			value = strings.TrimSpace(value)
			value = strings.TrimPrefix(value, "Scenario:")
			value = strings.TrimPrefix(value, "Scenario Outline:")
			value = strings.TrimSpace(value)
			if _, ok := resMap[value]; !ok {
				resMap[value] = true
			}
		}
		ret := make([]string, 0, len(resMap))
		for scenario := range resMap {
			ret = append(ret, "\""+scenario+"\"")
		}
		return ret
	}).([]string)
}
