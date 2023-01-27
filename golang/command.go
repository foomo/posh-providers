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
			return inst.completePaths(ctx, "go.mod")
		},
	}

	pathGenerateArg := &tree.Arg{
		Name:     "path",
		Optional: true,
		Suggest: func(ctx context.Context, p *tree.Root, r *readline.Readline) []prompt2.Suggest {
			return inst.completePaths(ctx, "generate.go")
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
						Args:        []*tree.Arg{pathModArg},
						Execute:     inst.modTidy,
					},
					{
						Name:        "download",
						Description: "run go mod download",
						Args:        []*tree.Arg{pathModArg},
						Execute:     inst.modDownload,
					},
					{
						Name:        "outdated",
						Description: "show go mod outdated",
						Args:        []*tree.Arg{pathModArg},
						Execute:     inst.modOutdated,
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
				Description: "run go mod download",
				Args:        []*tree.Arg{pathGenerateArg},
				Execute:     inst.generate,
			},
			// TODO test
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

func (c *Command) Help() string {
	return `Looks for go.mod files and runs the given command.

Usage:
  gomod [command] <path>

Available commands:
  tidy       run go mod tidy on specific or all paths
  download   run go mod download on specific or all paths
  outdated   list outdated packages on specific or all paths

Examples:
  gomod tidy ./path
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) modTidy(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx, "go.mod")
	}
	for _, value := range paths {
		c.l.Info("go mod tidy:", value)
		if err := shell.New(ctx, c.l,
			"go", "mod", "tidy",
		).
			Args(r.AdditionalArgs()...).
			Dir(path.Dir(value)).
			Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) modDownload(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx, "go.mod")
	}
	for _, value := range paths {
		c.l.Info("go mod download:", value)
		if err := shell.New(ctx, c.l,
			"go", "mod", "tidy",
		).
			Args(r.AdditionalArgs()...).
			Dir(path.Dir(value)).
			Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) modOutdated(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx, "go.mod")
	}
	for _, value := range paths {
		c.l.Info("go mod outdated:", value)
		if err := shell.New(ctx, c.l,
			"go", "list",
			"-u", "-m", "-json", "all",
			"|", "go-mod-outdated", "-update", "-direct",
		).
			Args(r.AdditionalArgs()...).
			Dir(value).
			Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) workInit(ctx context.Context, r *readline.Readline) error {
	data := "go 1.19\n\nuse (\n"
	for _, value := range c.paths(ctx, "go.mod") {
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
	if r.Args().HasIndex(2) {
		paths = append(paths, r.Args().At(2))
	} else {
		paths = c.paths(ctx, "generate.go")
	}

	for _, value := range paths {
		c.l.Info("go generate:", value)
		if err := shell.New(ctx, c.l,
			"go", "generate", value,
		).
			Args(r.AdditionalArgs()...).
			Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) completePaths(ctx context.Context, filename string) []goprompt.Suggest {
	return suggests.List(c.paths(ctx, filename))
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context, filename string) []string {
	return c.cache.Get("paths-"+filename, func() any {
		if value, err := files.Find(ctx, ".", filename); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else {
			return value
		}
	}).([]string)
}
