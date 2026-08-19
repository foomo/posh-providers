package k6

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/spf13/viper"
)

//go:embed SKILL.md
var skill string

type (
	Command struct {
		l           log.Logger
		op          *onepassword.OnePassword
		name        string
		cfg         Config
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

func NewCommand(l log.Logger, op *onepassword.OnePassword, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("k6"),
		name:      "k6",
		configKey: "k6",
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

	if err := os.MkdirAll(inst.cfg.Path, 0700); err != nil {
		return nil, err
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run a k6 load-test scenario against a configured environment",
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Default().Bool("verbose", false, "Enable k6 verbose logging")
			return nil
		},
		Args: tree.Args{
			{
				Name:        "env",
				Description: "Environment to run against, supplying its variables to the scenario",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.cfg.EnvNames())
				},
			},
			{
				Name:        "scenario",
				Description: "Scenario script to run, relative to the configured k6 path",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					root := env.Path(inst.cfg.Path)

					ret, err := files.Find(ctx, root, "*.k6.js", files.FindWithIsFile(true))
					if err != nil {
						return nil
					}

					for i, s := range ret {
						ret[i] = strings.TrimPrefix(s, root+"/")
					}

					return suggests.List(ret)
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

// Skill implements the optional command.Skiller interface. The catalog cannot
// show that a hardcoded --out web-dashboard keeps k6 alive while a browser is
// attached, that <env> selects the system under load and is never validated, or
// that op inject leaves decrypted secrets on disk beside the scenarios.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()
	env, scenario := r.Args().At(0), r.Args().At(1)

	var envs []string
	for k, v := range c.cfg.Env(env) {
		envs = append(envs, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}

	if c.op != nil {
		{
			secret := path.Join(c.cfg.Path, "all.k6.secret")
			if err := files.Exists(secret + ".tpl"); err == nil {
				if err := exec.CommandContext(ctx, "op", "inject", "-f", "-i", secret+".tpl", "-o", secret).Run(); err != nil {
					return err
				}
			}
		}
		{
			secret := path.Join(c.cfg.Path, strings.TrimSuffix(scenario, ".js")+".secret")
			if err := files.Exists(secret + ".tpl"); err == nil {
				if err := exec.CommandContext(ctx, "op", "inject", "-f", "-i", secret+".tpl", "-o", secret).Run(); err != nil {
					return err
				}
			}
		}
	}

	var args []string
	{
		secret := path.Join(c.cfg.Path, "all.k6.secret")
		if err := files.Exists(secret); err == nil {
			args = append(args, "--secret-source=file=name=all,filename="+secret)
		}
	}
	{
		secret := path.Join(c.cfg.Path, strings.TrimSuffix(scenario, ".js")+".secret")
		if err := files.Exists(secret); err == nil {
			args = append(args, "--secret-source=file=name=default,filename="+secret)
		}
	}

	return shell.New(ctx, c.l, "k6", "run").
		Args(args...).
		Args("--no-usage-report").
		Args("--out", "web-dashboard").
		Args(fs.Visited().Args()...).
		Args(path.Join(c.cfg.Path, scenario)).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Env(envs...).
		Run()
}
