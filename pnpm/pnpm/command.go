package pnpm

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

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
)

type (
	Command struct {
		l           log.Logger
		cache       cache.Namespace
		commandTree tree.Root
	}
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache) *Command {
	inst := &Command{
		l:     l.Named("pnpm"),
		cache: cache.Get("pnpm"),
	}

	fallback := func(ctx context.Context, r *readline.Readline) error {
		return shell.New(ctx, l, "pnpm").Args(r.Args()...).Args(r.Flags()...).Run()
	}
	workspaceFallback := func(ctx context.Context, r *readline.Readline) error {
		return shell.New(ctx, l, "pnpm").Dir(r.Args().At(1)).Args(r.Args().From(2)...).Args(r.Flags()...).Run()
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "pnpm",
		Description: "Fast, disk space efficient package manager",
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Default().Bool("recursive", false, "Run the command for each project in the workspace.")
			return nil
		},
		Nodes: tree.Nodes{
			{
				Name: "workspace",
				Nodes: tree.Nodes{
					{
						Name:        "path",
						Description: "Location to execute",
						Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.paths(ctx))
						},
						Nodes: tree.Nodes{
							{
								Name: "run",
								Args: tree.Args{
									{
										Name:        "script",
										Description: "Run scripts",
										Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
											return suggests.List(inst.scripts(ctx, r.Args().At(1)))
										},
									},
								},
								Execute: func(ctx context.Context, r *readline.Readline) error {
									return inst.run(ctx, r.Args().At(1), r.Args().At(3))
								},
							},
							{
								Name:        "add",
								Description: "Installs a package and any packages that it depends on",
								Execute:     workspaceFallback,
							},
							{
								Name:        "audit",
								Description: "Checks for known security issues with the installed packages",
								Execute:     workspaceFallback,
							},
							{
								Name:        "list",
								Description: "Print all the versions of packages that are installed, as well as their dependencies, in a tree-structure",
								Execute:     workspaceFallback,
							},
						},
						Execute: workspaceFallback,
					},
				},
			},
			{
				Name:        "install",
				Description: "Install dependencies",
				Execute: func(ctx context.Context, r *readline.Readline) error {
					return shell.New(ctx, l, "pnpm", "install").Run()
				},
			},
			{
				Name:        "run",
				Description: "Run script",
				Args: tree.Args{
					{
						Name:        "script",
						Description: "Run scripts",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.scripts(ctx, "."))
						},
					},
				},
				Execute: func(ctx context.Context, r *readline.Readline) error {
					return inst.run(ctx, ".", r.Args().At(1))
				},
			},
			{
				Name:        "add",
				Description: "Installs a package and any packages that it depends on",
				Execute:     fallback,
			},
			{
				Name:        "audit",
				Description: "Checks for known security issues with the installed packages",
				Execute:     fallback,
			},
			{
				Name:        "list",
				Description: "Print all the versions of packages that are installed, as well as their dependencies, in a tree-structure",
				Execute:     fallback,
			},
		},
		Execute: fallback,
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

func (c *Command) run(ctx context.Context, dirname, script string) error {
	fmt.Print(dirname, script)
	return shell.New(ctx, c.l, "pnpm", "run", script).
		Dir(dirname).
		Run()
}

//nolint:forcetypeassert
func (c *Command) scripts(ctx context.Context, dirname string) []string {
	return c.cache.Get("scripts-"+dirname, func() any {
		var ret []string
		f, err := LoadPackageJSON(path.Join(dirname, "package.json"))
		if err != nil {
			c.l.Debug("failed to load package.json", err.Error())
			return []string{}
		}

		for script := range f.Scripts {
			ret = append(ret, script)
		}

		return ret
	}).([]string)
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		var ret []string
		var roots []string

		{
			filename := env.Path("pnpm-workspace.yaml")
			if _, err := os.Stat(filename); errors.Is(err, os.ErrExist) {
				// do nothing
			} else if err != nil {
				c.l.Debug("failed to stat workspace file", err.Error())
				return []string{}
			} else {
				if value, err := LoadWorkspace(filename); err != nil {
					c.l.Debug("failed to stat workspace file", err.Error())
					return []string{}
				} else {
					for _, s := range value.Packages {
						roots = append(roots, strings.TrimSuffix(s, "/*"))
					}
				}
			}
		}

		if len(roots) == 0 {
			roots = append(roots, ".")
		}

		for _, root := range roots {
			if value, err := files.Find(ctx, root, "package.json", files.FindWithIgnore(`^(node_modules|\.\w*)$`)); err != nil {
				c.l.Debug("failed to walk files", err.Error())
				return []string{}
			} else {
				for _, s := range value {
					ret = append(ret, path.Dir(s))
				}
			}
		}

		return ret
	}).([]string)
}
