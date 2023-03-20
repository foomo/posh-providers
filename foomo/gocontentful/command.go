package gocontentful

import (
	"context"
	"os"
	"path"
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
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type (
	Command struct {
		l           log.Logger
		op          *onepassword.OnePassword
		name        string
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

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, op *onepassword.OnePassword, opts ...CommandOption) *Command {
	inst := &Command{
		l:     l.Named("gocontentful"),
		name:  "gocontentful",
		cache: cache.Get("gocontentful"),
		op:    op,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}
	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run gocontentful",
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Default().Bool("debug", false, "show debug output")
			return nil
		},
		Args: tree.Args{
			{
				Name:     "path",
				Optional: true,
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.paths(ctx))
				},
			},
		},
		Execute: inst.execute,
	})

	return inst
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
		return nil
	case r.Args().LenGt(1):
		return errors.New("too many arguments")
	}

	if info, err := os.Stat(r.Args().At(0)); err != nil || info.IsDir() {
		return errors.New("invalid [path] parameter")
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
	var paths []string
	if r.Args().HasIndex(0) {
		paths = []string{r.Args().At(0)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running gocontentful...")
	for _, value := range paths {
		c.l.Info("└  " + value)

		var cfg Config
		if out, err := c.op.RenderFile(ctx, value); err != nil {
			return errors.Wrap(err, "failed to render gocontentful.yaml")
		} else if err := yaml.Unmarshal(out, &cfg); err != nil {
			return errors.Wrap(err, "failed to unmarshal gocontentful.yaml")
		}

		dir := path.Dir(value)
		if out, err := shell.New(ctx, c.l, "rm", "-f", dir+"/gocontentful*.go").Output(); err != nil {
			return errors.Wrap(err, string(out))
		}

		if err := shell.New(ctx, c.l, "gocontentful",
			"-spaceid", cfg.SpaceID, "-cmakey", cfg.CMAKey,
			"-contenttypes", strings.Join(cfg.ContentTypes, ","), dir).
			Args(r.AdditionalArgs()...).
			Run(); err != nil {
			return err
		}
	}
	return nil
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		if value, err := files.Find(ctx, ".", "gocontentful.yml"); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else {
			return value
		}
	}).([]string)
}
