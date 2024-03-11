package sesamy

import (
	"bytes"
	"context"
	"path"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/prints"
	"github.com/foomo/posh/pkg/util/suggests"
)

type Command struct {
	l           log.Logger
	op          *onepassword.OnePassword
	cache       cache.Namespace
	commandTree tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, op *onepassword.OnePassword, cache cache.Cache) *Command {
	inst := &Command{
		l:  l.Named("sesamy"),
		op: op,

		cache: cache.Get("sesamy"),
	}

	flags := func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
		fs.Default().Bool("verbose", false, "show verbose output")
		return nil
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "sesamy",
		Description: "Run sesamy",
		Nodes: tree.Nodes{
			{
				Name:        "config",
				Description: "Dump config",
				Args: tree.Args{
					{
						Name:     "path",
						Optional: true,
						Suggest:  inst.completePaths,
					},
				},
				Execute: inst.config,
			},
			{
				Name:        "typescript",
				Description: "Generate typescript definitions",
				Flags:       flags,
				Args: tree.Args{
					{
						Name:     "path",
						Optional: true,
						Suggest:  inst.completePaths,
					},
				},
				Execute: inst.typescript,
			},
			{
				Name:        "tagmanager",
				Description: "Provision google tag manager",
				Nodes: tree.Nodes{
					{
						Name:        "web",
						Description: "Provision web container",
						Flags:       flags,
						Execute:     inst.tagmanagerWeb,
					},
					{
						Name:        "server",
						Description: "Provision server container",
						Flags:       flags,
						Execute:     inst.tagmanagerServer,
					},
				},
			},
		},
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

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) config(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Printing sesamy config ...")
	for _, value := range paths {
		c.l.Info("└ " + value)

		out, err := c.op.RenderFile(ctx, value)
		if err != nil {
			return err
		}

		prints.Code(c.l, value, string(out), "yaml")
	}
	return nil
}

func (c *Command) typescript(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy typescript ...")
	for _, value := range paths {
		c.l.Info("└ " + value)

		out, err := c.op.RenderFile(ctx, value)
		if err != nil {
			return err
		}

		if err := shell.New(ctx, c.l, "sesamy", "typescript").
			Args(r.Flags()...).
			Args("--config", "-").
			Args(r.AdditionalArgs()...).
			Stdin(bytes.NewReader(out)).
			Dir(path.Dir(value)).
			Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) tagmanagerWeb(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy tagmanger web ...")
	for _, value := range paths {
		c.l.Info("└ " + value)

		out, err := c.op.RenderFile(ctx, value)
		if err != nil {
			return err
		}

		if err := shell.New(ctx, c.l, "sesamy", "tagmanager", "web").
			Args(r.Flags()...).
			Args("--config", "-").
			Args(r.AdditionalArgs()...).
			Stdin(bytes.NewReader(out)).
			Dir(path.Dir(value)).
			Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) tagmanagerServer(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy tagmanger server ...")
	for _, value := range paths {
		c.l.Info("└ " + value)

		out, err := c.op.RenderFile(ctx, value)
		if err != nil {
			return err
		}

		if err := shell.New(ctx, c.l, "sesamy", "tagmanager", "server").
			Args(r.Flags()...).
			Args("--config", "-").
			Args(r.AdditionalArgs()...).
			Stdin(bytes.NewReader(out)).
			Dir(path.Dir(value)).
			Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) completePaths(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	return suggests.List(c.paths(ctx))
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		if value, err := files.Find(ctx, ".", "sesamy.yml"); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return nil
		} else {
			return value
		}
	}).([]string)
}
