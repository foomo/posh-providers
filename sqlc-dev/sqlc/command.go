package sqlc

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
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
		cache       cache.Namespace
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

func NewCommand(l log.Logger, cache cache.Cache, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		name:  "sqlc",
		l:     l.Named("sqlc"),
		cache: cache.Get("sqlc"),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}
	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	pathArgs := tree.Args{
		{
			Name:     "path",
			Optional: true,
			Suggest:  inst.completePaths,
		},
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run sqlc",
		Nodes: []*tree.Node{
			{
				Name:        "compile",
				Args:        pathArgs,
				Description: "Statically check SQL for syntax and type errors",
				Execute:     inst.run,
			},
			{
				Name:        "generate",
				Args:        pathArgs,
				Description: "Generate source code from SQL",
				Execute:     inst.run,
			},
			{
				Name:        "diff",
				Args:        pathArgs,
				Description: "Statically check SQL for syntax and type errors",
				Execute:     inst.run,
			},
			{
				Name:        "vet",
				Args:        pathArgs,
				Description: "Statically check SQL for syntax and type errors",
				Execute:     inst.run,
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

func (c *Command) run(ctx context.Context, r *readline.Readline) error {
	cmd := r.Args().At(0)

	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sqlc " + cmd + " ...")
	for _, value := range paths {
		c.l.Info("└ " + value)
		if out, err := shell.New(ctx, c.l, "sqlc", "--no-remote", cmd).
			Dir(path.Dir(value)).
			Args(r.AdditionalArgs()...).
			Env("SQLCCACHE=" + env.Path(c.cfg.CacheDir)).
			Env("SQLCTMPDIR=" + env.Path(c.cfg.TempDir)).
			Output(); err != nil {
			return errors.Wrap(err, string(out))
		}
	}
	return nil
}

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "sqlc", "--no-remote").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Env("SQLCCACHE=" + env.Path(c.cfg.CacheDir)).
		Env("SQLCTMPDIR=" + env.Path(c.cfg.TempDir)).
		Run()
}

func (c *Command) completePaths(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	return suggests.List(c.paths(ctx))
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		if value, err := files.Find(ctx, ".", "sqlc.yaml"); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return nil
		} else {
			return value
		}
	}).([]string)
}
