package golang

import (
	"context"
	"os"
	"path"
	"strings"

	prompt2 "github.com/c-bata/go-prompt"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"golang.org/x/sync/errgroup"
)

type Command struct {
	l           log.Logger
	cache       cache.Namespace
	commandTree *tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache) *Command {
	inst := &Command{
		l:     l.Named("go"),
		cache: cache.Get("go"),
	}

	pathModArg := &tree.Arg{
		Name:     "path",
		Optional: true,
		Suggest: func(ctx context.Context, p *tree.Root, r *readline.Readline) []prompt2.Suggest {
			return inst.completePaths(ctx, "go.mod", true)
		},
	}

	pathGenerateArg := &tree.Arg{
		Name:     "path",
		Optional: true,
		Suggest: func(ctx context.Context, p *tree.Root, r *readline.Readline) []prompt2.Suggest {
			return inst.completePaths(ctx, "generate.go", false)
		},
	}

	inst.commandTree = &tree.Root{
		Name:        "go",
		Description: "go related tasks",
		Nodes: tree.Nodes{
			{
				Name:        "mod",
				Description: "run go mod commands",
				Nodes: tree.Nodes{
					{
						Name:        "tidy",
						Description: "run go mod tidy",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							fs.Int("parallel", 0, "number of parallel processes")
							return nil
						},
						Args:    []*tree.Arg{pathModArg},
						Execute: inst.modTidy,
					},
					{
						Name:        "download",
						Description: "run go mod download",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							fs.Int("parallel", 0, "number of parallel processes")
							return nil
						},
						Args:    []*tree.Arg{pathModArg},
						Execute: inst.modDownload,
					},
					{
						Name:        "outdated",
						Description: "show go mod outdated",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							fs.Int("parallel", 0, "number of parallel processes")
							return nil
						},
						Args:    []*tree.Arg{pathModArg},
						Execute: inst.modOutdated,
					},
				},
			},
			{
				Name:        "work",
				Description: "manage go.work file",
				Nodes: tree.Nodes{
					{
						Name:        "init",
						Description: "generate go.work file",
						Execute:     inst.workInit,
					},
					{
						Name:        "use",
						Description: "add go.work entry",
						Args: []*tree.Arg{
							{
								Name:    "path",
								Suggest: nil,
							},
						},
						Execute: inst.workUse,
					},
				},
			},
			{
				Name:        "generate",
				Description: "run go mod commands",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
					fs.Int("parallel", 0, "number of parallel processes")
					return nil
				},
				Args:    []*tree.Arg{pathGenerateArg},
				Execute: inst.generate,
			},
			{
				Name:        "test",
				Description: "run go test",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
					fs.Int("parallel", 0, "number of parallel processes")
					return nil
				},
				Args:    []*tree.Arg{pathModArg},
				Execute: inst.test,
			},
			{
				Name:        "build",
				Description: "run go build",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
					fs.Int("parallel", 0, "number of parallel processes")
					return nil
				},
				Args:    []*tree.Arg{pathModArg},
				Execute: inst.build,
			},
		},
	}
	return inst
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
	return `Looks for go.mod files and runs the given command.

Usage:
  go [command]

Available commands:
  mod [command]     run go mod
  generate <path>   run go generate
  test <path>       run go test
  build <path>      run go build

SubCommands mod:
  tidy <path>
  download <path>
  outdated <path>
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) build(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}

	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go build ...")
	for _, value := range paths {
		value := value
		wg.Go(func() error {
			c.l.Info("└ " + value)
			return shell.New(ctx, c.l,
				"go", "build", "-v", "./...", // TODO select test
			).
				Args(r.PassThroughFlags()...).
				Args(r.AdditionalArgs()...).
				Dir(value).
				Run()
		})
	}
	return wg.Wait()
}

func (c *Command) test(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}

	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go test ...")
	for _, value := range paths {
		value := value
		wg.Go(func() error {
			c.l.Info("└ " + value)
			return shell.New(ctx, c.l,
				"go", "test", "-v", "./...", // TODO select test
			).
				Args(r.PassThroughFlags()...).
				Args(r.AdditionalArgs()...).
				Dir(value).
				Run()
		})
	}
	return wg.Wait()
}

func (c *Command) modTidy(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}
	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go mod tidy...")
	for _, value := range paths {
		value := value
		wg.Go(func() error {
			c.l.Info("└ " + value)
			return shell.New(ctx, c.l,
				"go", "mod", "tidy",
			).
				Args(r.AdditionalArgs()...).
				Dir(value).
				Run()
		})
	}
	return wg.Wait()
}

func (c *Command) modDownload(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}
	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go mod download...")
	for _, value := range paths {
		value := value
		wg.Go(func() error {
			c.l.Info("└ " + value)
			return shell.New(ctx, c.l,
				"go", "mod", "tidy",
			).
				Args(r.AdditionalArgs()...).
				Dir(value).
				Run()
		})
	}
	return wg.Wait()
}

func (c *Command) modOutdated(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx, "go.mod", true)
	}
	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go mod outdated...")
	for _, value := range paths {
		value := value
		wg.Go(func() error {
			c.l.Info("└ " + value)
			return shell.New(ctx, c.l,
				"go", "list",
				"-u", "-m", "-json", "all",
				"|", "go-mod-outdated", "-update", "-direct",
			).
				Args(r.AdditionalArgs()...).
				Dir(value).
				Run()
		})
	}
	return wg.Wait()
}

func (c *Command) workInit(ctx context.Context, r *readline.Readline) error {
	data := "go 1.19\n\nuse (\n"
	for _, value := range c.paths(ctx, "go.mod", true) {
		data += "\t" + strings.TrimSuffix(value, "/go.mod") + "\n"
	}
	data += ")"
	return os.WriteFile(path.Join(os.Getenv("PROJECT_ROOT"), "go.work"), []byte(data), 0600)
}

func (c *Command) workUse(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "go").
		Args(r.Args()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) generate(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = append(paths, r.Args().At(1))
	} else {
		paths = c.paths(ctx, "generate.go", false)
	}

	ctx, wg := c.wg(ctx, r)
	c.l.Info("Running go generate...")
	for _, value := range paths {
		value := value
		wg.Go(func() error {
			c.l.Info("└ " + value)
			return shell.New(ctx, c.l,
				"go", "generate", value,
			).
				Args(r.AdditionalArgs()...).
				Run()
		})
	}

	return wg.Wait()
}

func (c *Command) completePaths(ctx context.Context, filename string, dir bool) []goprompt.Suggest {
	return suggests.List(c.paths(ctx, filename, dir))
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context, filename string, dir bool) []string {
	return c.cache.Get("paths-"+filename, func() any {
		if value, err := files.Find(ctx, ".", filename); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else if dir {
			for i, s := range value {
				value[i] = path.Dir(s)
			}
			return value
		} else {
			return value
		}
	}).([]string)
}

func (c *Command) wg(ctx context.Context, r *readline.Readline) (context.Context, *errgroup.Group) {
	wg, ctx := errgroup.WithContext(ctx)
	if value, err := r.FlagSet().GetInt("parallel"); err == nil && value != 0 {
		wg.SetLimit(value)
	} else {
		wg.SetLimit(1)
	}
	return ctx, wg
}
