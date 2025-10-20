package hygen

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
		cfg         Config
		name        string
		cache       cache.Namespace
		configKey   string
		commandTree tree.Root
	}
	Option func(*Command) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) Option {
	return func(o *Command) error {
		o.name = v
		return nil
	}
}

func WithConfigKey(v string) Option {
	return func(o *Command) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, opts ...Option) (*Command, error) {
	inst := &Command{
		l:         l.Named("hygen"),
		name:      "hygen",
		configKey: "hygen",
		cache:     cache.Get("hygen"),
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(inst); err != nil {
				return nil, err
			}
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run hygen",
		Args: tree.Args{
			{
				Name: "path",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.paths(ctx))
				},
			},
		},
		Nodes: tree.Nodes{
			{
				Name:        "template",
				Description: "Render template",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.paths(ctx))
				},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Default().Bool("dry", false, "Perform a dry run. Files will be generated but not saved")
					return nil
				},
				Args: tree.Args{
					{
						Name: "path",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return nil
						},
					},
				},
				Execute: inst.execute,
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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	switch {
	case r.Args().LenIs(0):
		return errors.New("missing [template] argument")
	case r.Args().LenIs(1):
		return errors.New("missing [path] argument")
	case r.Args().LenGt(2):
		return errors.New("too many arguments")
	}

	if info, err := os.Stat(filepath.Join(c.cfg.TemplatePath, r.Args().At(0))); err != nil || !info.IsDir() {
		return errors.Errorf("invalid [TEMPLATE] parameter: %s", r.Args().At(0))
	}

	return nil
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
	return shell.New(ctx, c.l, "hygen", "scaffold").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Env(fmt.Sprintf("HYGEN_TMPLS=%s", path.Dir(c.cfg.TemplatePath))).
		Run()
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		files, err := os.ReadDir(c.cfg.TemplatePath)
		if err != nil {
			c.l.Debug("failed to read template dir:", err.Error())
			return []string{}
		}

		ret := make([]string, 0, len(files))
		for _, value := range files {
			if value.IsDir() {
				ret = append(ret, value.Name())
			}
		}

		return ret
	}).([]string)
}
