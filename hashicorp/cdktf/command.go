package cdktf

import (
	"context"
	"errors"
	"os"
	"os/exec"
	path2 "path"
	"strings"

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
		l                 log.Logger
		cfg               Config
		name              string
		cache             cache.Namespace
		configKey         string
		commandTree       tree.Root
		stackNameProvider StackNameProvider
	}
	StackNameProvider func(path string) string
	CommandOption     func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

func CommandWithStackNameProvider(v StackNameProvider) CommandOption {
	return func(o *Command) {
		o.stackNameProvider = v
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

func NewCommand(l log.Logger, cache cache.Cache, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("cdktf"),
		name:      "cdktf",
		configKey: "cdktf",
		cache:     cache.Get("cdktf"),
		stackNameProvider: func(path string) string {
			return strings.TrimSuffix(path2.Base(path), ".stack.yaml")
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	if l.IsLevel(log.LevelTrace) {
		_ = os.Setenv("CDKTF_LOG_LEVEL", "debug")
	} else if l.IsLevel(log.LevelDebug) {
		_ = os.Setenv("CDKTF_LOG_LEVEL", "info")
	}

	stacksArg := &tree.Arg{
		Name:        "stacks",
		Description: "Name of the stacks",
		Repeat:      true,
		Optional:    true,
		Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
			return suggests.List(inst.stacks(ctx))
		},
	}
	skipSynthFlag := func(fs *readline.FlagSets) {
		fs.Internal().Bool("skip-synth", false, "Skip synth trough env var")
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run cdktf",
		Nodes: tree.Nodes{
			{
				Name:        "list",
				Description: "List stacks in app",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					skipSynthFlag(fs)
					return nil
				},
				Execute: inst.list,
			},
			{
				Name:        "diff",
				Description: "Perform a diff (terraform plan) for the given stack",
				Args:        tree.Args{stacksArg},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					skipSynthFlag(fs)
					fs.Default().Bool("refresh-only", false, "Select the 'refresh only' planning mode")
					fs.Default().Bool("migrate-state", false, "Pass this flag after switching state backends")

					return nil
				},
				Execute: inst.diff,
			},
			{
				Name:        "deploy",
				Description: "Deploy the given stacks",
				Args:        tree.Args{stacksArg},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					skipSynthFlag(fs)
					fs.Default().Bool("auto-approve", false, "Auto approve")
					fs.Default().Bool("refresh-only", false, "Select the 'refresh only' planning mode")
					fs.Default().Bool("migrate-state", false, "Pass this flag after switching state backends")

					return nil
				},
				Execute: inst.deploy,
			},
			{
				Name:        "destroy",
				Description: "Destroy the given stacks",
				Args:        tree.Args{stacksArg},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					skipSynthFlag(fs)
					fs.Default().Bool("auto-approve", false, "Auto approve")
					fs.Default().Bool("migrate-state", false, "Pass this flag after switching state backends")

					return nil
				},
				Execute: inst.destroy,
			},
			{
				Name:        "unlock",
				Description: "Unlock a terraform state.",
				Args: tree.Args{
					{
						Name:        "stack",
						Description: "Path to the terraform stack.",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.stacks(ctx))
						},
					},
					{
						Name:        "lockId",
						Description: "Terraform stage lock id",
					},
				},
				Execute: inst.unlock,
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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	if _, err := exec.LookPath("cdktf"); err != nil {
		c.l.Print()

		return errors.New(`
Please ensure you have the cdktf installed!

- Install binary:

$ npm install --global cdktf-cli
        `)
	}

	return nil
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) list(ctx context.Context, r *readline.Readline) error {
	var envs []string

	ifs := r.FlagSets().Internal()
	if value, err := ifs.GetBool("skip-synth"); err != nil {
		return err
	} else if value {
		envs = append(envs, "SKIP_SYNTH=true")
	}

	return shell.New(ctx, c.l, "cdktf", "list").
		Env(envs...).
		Dir(c.cfg.Path).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) diff(ctx context.Context, r *readline.Readline) error {
	var envs []string

	ifs := r.FlagSets().Internal()
	if value, err := ifs.GetBool("skip-synth"); err != nil {
		return err
	} else if value {
		envs = append(envs, "SKIP_SYNTH=true")
	}

	return shell.New(ctx, c.l, "cdktf", "diff").
		Env(envs...).
		Dir(c.cfg.Path).
		Args(r.Args().From(1)...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) deploy(ctx context.Context, r *readline.Readline) error {
	var envs []string

	fs := r.FlagSets().Default()

	ifs := r.FlagSets().Internal()
	if value, err := ifs.GetBool("skip-synth"); err != nil {
		return err
	} else if value {
		envs = append(envs, "SKIP_SYNTH=true")
	}

	return shell.New(ctx, c.l, "cdktf", "deploy").
		Env(envs...).
		Dir(c.cfg.Path).
		Args(r.Args().From(1)...).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) destroy(ctx context.Context, r *readline.Readline) error {
	var envs []string

	ifs := r.FlagSets().Internal()
	if value, err := ifs.GetBool("skip-synth"); err != nil {
		return err
	} else if value {
		envs = append(envs, "SKIP_SYNTH=true")
	}

	return shell.New(ctx, c.l, "cdktf", "destroy").
		Env(envs...).
		Dir(c.cfg.Path).
		Args(r.Args().From(1)...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) unlock(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "terraform", "force-unlock", r.Args().At(2)).
		Dir(path2.Join(c.cfg.OutPath, "stacks", r.Args().At(1))).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

//nolint:forcetypeassert
func (c *Command) stacks(ctx context.Context) []string {
	return c.cache.Get("stacks", func() any {
		if value, err := files.Find(ctx, c.cfg.Path, "*.stack.yaml", files.FindWithIgnore(`^\.`)); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else {
			for i, s := range value {
				value[i] = c.stackNameProvider(s)
			}

			return value
		}
	}).([]string)
}
