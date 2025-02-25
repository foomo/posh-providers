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

	configArg := &tree.Arg{
		Name:     "path",
		Optional: true,
		Suggest:  inst.completePaths,
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
				Args:        tree.Args{configArg},
				Execute:     inst.config,
			},
			{
				Name:        "tags",
				Description: "List available tags",
				Args:        tree.Args{configArg},
				Flags:       flags,
				Execute:     inst.tags,
			},
			{
				Name:        "typescript",
				Description: "Generate typescript definitions",
				Args:        tree.Args{configArg},
				Flags:       flags,
				Execute:     inst.typescript,
			},
			{
				Name:        "provision",
				Description: "Provision Google Tag Manager",
				Nodes: tree.Nodes{
					{
						Name:        "web",
						Description: "Provision Web Container",
						Args:        tree.Args{configArg},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().StringSlice("tags", nil, "list of tags to run")
							return flags(ctx, r, fs)
						},
						Execute: inst.provisionWeb,
					},
					{
						Name:        "server",
						Description: "Provision Server Container",
						Args:        tree.Args{configArg},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().StringSlice("tags", nil, "list of tags to run")
							return flags(ctx, r, fs)
						},
						Execute: inst.provisionServer,
					},
				},
			},
			{
				Name:        "list",
				Description: "List Google Tag Manager Resources",
				Nodes: tree.Nodes{
					{
						Name:        "web",
						Description: "List Web Container Resources",
						Args: tree.Args{
							{
								Name:        "resource",
								Description: "Name of the resource",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return []goprompt.Suggest{
										{Text: "built-in-variables"},
										{Text: "folders"},
										{Text: "gtag-config"},
										{Text: "status"},
										{Text: "tags"},
										{Text: "templates"},
										{Text: "templates-data"},
										{Text: "transformations"},
										{Text: "triggers"},
										{Text: "variables"},
										{Text: "zones"},
									}
								},
							},
							configArg,
						},
						Execute: inst.listWeb,
					},
					{
						Name:        "server",
						Description: "List Server Container Resources",
						Args: tree.Args{
							{
								Name:        "resource",
								Description: "Name of the resource",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return []goprompt.Suggest{
										{Text: "built-in-variables"},
										{Text: "clients"},
										{Text: "folders"},
										{Text: "gtag-config"},
										{Text: "status"},
										{Text: "tags"},
										{Text: "templates"},
										{Text: "templates-data"},
										{Text: "transformations"},
										{Text: "triggers"},
										{Text: "variables"},
										{Text: "zones"},
									}
								},
							},
							configArg,
						},
						Execute: inst.listServer,
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

		if err := shell.New(ctx, c.l, "sesamy", "config").
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

func (c *Command) tags(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy tags ...")
	for _, value := range paths {
		c.l.Info("└ " + value)

		out, err := c.op.RenderFile(ctx, value)
		if err != nil {
			return err
		}

		if err := shell.New(ctx, c.l, "sesamy", "tags").
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

func (c *Command) provision(ctx context.Context, r *readline.Readline, cmd string) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy provision " + cmd + " ...")
	for _, value := range paths {
		c.l.Info("└ " + value)

		out, err := c.op.RenderFile(ctx, value)
		if err != nil {
			return err
		}

		if err := shell.New(ctx, c.l, "sesamy", "provision", cmd).
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

func (c *Command) provisionWeb(ctx context.Context, r *readline.Readline) error {
	return c.provision(ctx, r, "web")
}

func (c *Command) provisionServer(ctx context.Context, r *readline.Readline) error {
	return c.provision(ctx, r, "server")
}

func (c *Command) listWeb(ctx context.Context, r *readline.Readline) error {
	return c.list(ctx, r, "web")
}

func (c *Command) listServer(ctx context.Context, r *readline.Readline) error {
	return c.list(ctx, r, "server")
}

func (c *Command) list(ctx context.Context, r *readline.Readline, cmd string) error {
	resource := r.Args().At(2)

	var paths []string
	if r.Args().HasIndex(3) {
		paths = []string{r.Args().At(3)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy list " + cmd + " ...")
	for _, value := range paths {
		c.l.Info("└ " + value)

		out, err := c.op.RenderFile(ctx, value)
		if err != nil {
			return err
		}

		if err := shell.New(ctx, c.l, "sesamy", "list", cmd, resource).
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
		if value, err := files.Find(ctx, ".", "sesamy*.yml"); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return nil
		} else {
			return value
		}
	}).([]string)
}
