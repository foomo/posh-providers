package mjml

import (
	"context"
	"os"
	"strings"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type (
	Command struct {
		l           log.Logger
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

func NewCommand(l log.Logger, cache cache.Cache, opts ...CommandOption) *Command {
	inst := &Command{
		l:     l.Named("mjml"),
		name:  "mjml",
		cache: cache.Get("mjml"),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run mjml",
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Internal().Int("parallel", 0, "number of parallel processes")
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

	if info, err := os.Stat(r.Args().At(0)); err != nil || !info.IsDir() {
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
	dir := "."
	if r.Args().HasIndex(0) {
		dir = r.Args().At(0)
	}

	ctx, wg := c.wg(ctx, r)
	c.l.Infof("Running mjml under %q", dir)

	for _, src := range c.files(ctx, dir) {
		wg.Go(func() error {
			c.l.Info("└  " + src)
			out := strings.ReplaceAll(src, ".mjml", ".html")
			out = strings.ReplaceAll(out, "/src/", "/html/")

			return shell.New(ctx, c.l, "mjml", src, "-o", out).
				Args(r.AdditionalArgs()...).
				Run()
		})
	}

	return wg.Wait()
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		if value, err := files.Find(ctx, ".", "*.mjml"); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else {
			ret := make([]string, 0, len(value))
			for _, s := range value {
				if v := strings.Split(s, "/src/"); len(v) == 2 {
					ret = append(ret, v[0])
				}
			}

			return lo.Uniq(ret)
		}
	}).([]string)
}

//nolint:forcetypeassert
func (c *Command) files(ctx context.Context, root string) []string {
	cacheKey := "files"
	if value := strings.TrimPrefix(root, "."); value != "" {
		cacheKey += strings.ReplaceAll(value, "/", "-")
	}

	return c.cache.Get(cacheKey, func() any {
		if value, err := files.Find(ctx, ".", "*.mjml"); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else {
			ret := make([]string, 0, len(value))
			for _, s := range value {
				if strings.Contains(s, "/src/") {
					ret = append(ret, s)
				}
			}

			return ret
		}
	}).([]string)
}

func (c *Command) wg(ctx context.Context, r *readline.Readline) (context.Context, *errgroup.Group) {
	wg, ctx := errgroup.WithContext(ctx)
	if value, _ := r.FlagSets().Internal().GetInt("parallel"); value != 0 {
		wg.SetLimit(value)
	} else {
		wg.SetLimit(1)
	}

	return ctx, wg
}
