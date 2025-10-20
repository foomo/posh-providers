package bun

import (
	"context"
	"path"
	"strings"

	"github.com/foomo/posh-providers/pkg/npm"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
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
		l:     l.Named("bun"),
		cache: cache.Get("bun"),
	}

	globalFlags := func(child func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error) func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
		return func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Default().Bool("watch", false, "Automatically restart the process on file change")
			fs.Default().Bool("hot", false, "Enable auto reload in the Bun runtime, test runner, or bundler")
			fs.Default().Bool("smol", false, "Use less memory, but run garbage collection more often")
			fs.Default().Bool("preload", false, "Import a module before other modules are loaded")
			fs.Default().Bool("inspect", false, "Activate Bun's debugger")
			fs.Default().Bool("inspect-wait", false, "Activate Bun's debugger, wait for a connection before executing")
			fs.Default().Bool("inspect-brk", false, "Activate Bun's debugger, set breakpoint on first line of code and wait")
			fs.Default().Bool("if-present", false, "Exit without an error if the entrypoint does not exist")
			fs.Default().Bool("no-install", false, "Disable auto install in the Bun runtime")
			fs.Default().Bool("install", false, "Configure auto-install behavior")
			// fs.Default().Bool("eval", false, "Evaluate argument as a script")
			fs.Default().Bool("print", false, "Evaluate argument as a script and print the result")
			fs.Default().Bool("prefer-offline", false, "Skip staleness checks for packages in the Bun runtime and resolve from disk")
			fs.Default().Bool("prefer-latest", false, "Use the latest matching versions of packages in the Bun runtime, always checking npm")
			// fs.Default().Bool("port", false, "Set the default port for Bun.serve")
			// fs.Default().Bool("conditions", false, "Pass custom conditions to resolve")
			// fs.Default().Bool("fetch-preconnect", false, "Preconnect to a URL while code is loading")
			// fs.Default().Bool("max-http-header-size", false, "Set the maximum size of HTTP headers in bytes. Default is 16KiB")
			// fs.Default().Bool("expose-internals", false, "Expose internals used for testing Bun itself. Usage of these APIs are completely unsupported.")
			// fs.Default().Bool("expose-gc", false, "Expose gc() on the global object. Has no effect on Bun.gc().")
			// fs.Default().Bool("no-deprecation", false, "Suppress all reporting of the custom deprecation.")
			// fs.Default().Bool("throw-deprecation", false, "Determine whether or not deprecation warnings result in errors.")
			// fs.Default().Bool("title", false, "Set the process title")
			// fs.Default().Bool("silent", false, "Don't print the script command")
			fs.Default().String("elide-lines", "", "Number of lines of script output shown when using --filter (default: 10). Set to 0 to show all lines.")
			fs.Default().String("filter", "", "Run a script in all workspace packages matching the pattern")
			fs.Default().Bool("bun", false, "Force a script or package to use Bun's runtime instead of Node.js (via symlinking node)")
			fs.Default().String("env-file", "", "Load environment variables from the specified file(s)")
			// fs.Default().String("cwd", "", "Absolute path to resolve files & entry points from. This just changes the process' cwd.")
			// fs.Default().String("config", "", "Specify path to Bun config file. Default $cwd/bunfig.toml")
			if child == nil {
				return nil
			}

			return child(ctx, r, fs)
		}
	}
	installFlags := func(child func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error) func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
		return func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Default().Bool("production", false, "Don't install devDependencie")
			fs.Default().Bool("save", false, "Save to package.json")
			fs.Default().Bool("omit", false, "Exclude 'dev', 'optional', or 'peer' dependencies from install")
			fs.Default().Bool("save-text-lockfile", false, " Save a text-based lockfile")
			fs.Default().Bool("lockfile-only", false, " Save a text-based lockfile")

			if child == nil {
				return nil
			}

			return child(ctx, r, fs)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "bun",
		Description: "All-in-one toolkit for JavaScript and TypeScript apps",
		Flags:       globalFlags(nil),
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
								Name:        "run",
								Description: "Execute a file with Bun",
								Args: tree.Args{
									{
										Name:        "script",
										Description: "Run scripts",
										Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
											return suggests.List(inst.scripts(ctx, r.Args().At(1)))
										},
									},
								},
								Flags:   globalFlags(nil),
								Execute: inst.runWorkspace,
							},
							{
								Name:        "test",
								Description: "Run unit tests with Bun",
								Flags: globalFlags(func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
									fs.Default().Bool("update-snapshots", false, "Update snapshot files")
									fs.Default().Bool("only", false, "Only run tests that are marked with test.only()")
									fs.Default().Bool("todo", false, "Include tests that are marked with test.todo()")
									fs.Default().String("timeout", "", "Set the per-test timeout in milliseconds, default is 5000")
									fs.Default().String("bail", "", "Exit the test suite after <NUMBER> failures")
									fs.Default().String("test-name-pattern", "", "Run only tests with a name that matches the given regex")

									return nil
								}),
								Execute: inst.runWorkspace,
							},
							{
								Name:        "x",
								Description: "Execute a package binary (CLI), installing if needed",
								Flags:       globalFlags(nil),
								Execute:     inst.runWorkspace,
							},
							{
								Name:        "add",
								Description: "Installs a package and any packages that it depends on",
								Flags: globalFlags(installFlags(func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
									fs.Default().Bool("dev", false, "Add dependency to devDependencies")
									fs.Default().Bool("peer", false, "Add dependency to peerDependencies")
									fs.Default().Bool("optional", false, "Add dependency to optionalDependencies")
									fs.Default().Bool("exact", false, "Add the exact version instead of the ^range")

									return nil
								})),
								Execute: inst.runWorkspace,
							},
							{
								Name:        "install",
								Description: "Install dependencies for a package.json",
								Flags:       globalFlags(installFlags(nil)),
								Execute:     inst.runWorkspace,
							},
							{
								Name:        "remove",
								Description: "Remove a dependency from package.json",
								Flags:       globalFlags(nil),
								Execute:     inst.runWorkspace,
							},
							{
								Name:        "update",
								Description: "Update outdated dependencies",
								Flags:       globalFlags(installFlags(nil)),
								Execute:     inst.runWorkspace,
							},
							{
								Name:        "patch",
								Description: "Prepare a package for patchin",
								Flags:       globalFlags(nil),
								Execute:     inst.runWorkspace,
							},
							{
								Name:        "outdated",
								Description: "Display latest versions of outdated dependencies",
								Flags:       globalFlags(nil),
								Execute:     inst.runWorkspace,
							},
							{
								Name:        "pm",
								Description: "Additional package management utilities",
								Args: tree.Args{
									{
										Name:        "subcommand",
										Description: "Package command",
									},
								},
								Flags:   globalFlags(nil),
								Execute: inst.runWorkspace,
							},
						},
						Execute: inst.runWorkspace,
					},
				},
			},
			{
				Name:        "x",
				Description: "Execute a package binary (CLI), installing if needed",
				Flags:       globalFlags(nil),
				Execute:     inst.run,
			},
			{
				Name:        "run",
				Description: "Execute a file with Bun",
				Args: tree.Args{
					{
						Name:        "script",
						Description: "Run scripts",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.scripts(ctx, "."))
						},
					},
				},
				Flags:   globalFlags(nil),
				Execute: inst.run,
			},
			{
				Name:        "install",
				Description: "Install dependencies for a package.json",
				Args: tree.Args{
					{
						Name:        "package",
						Optional:    true,
						Description: "Package name",
					},
				},
				Flags:   globalFlags(nil),
				Execute: inst.run,
			},
			{
				Name:        "remove",
				Description: "Remove a dependency from package.json",
				Args: tree.Args{
					{
						Name:        "package",
						Description: "Package name",
					},
				},
				Flags:   globalFlags(nil),
				Execute: inst.run,
			},
			{
				Name:        "update",
				Description: "Update outdated dependencies",
				Args: tree.Args{
					{
						Name:        "package",
						Description: "Package name",
					},
				},
				Flags:   globalFlags(nil),
				Execute: inst.run,
			},
			{
				Name: "patch",
				Args: tree.Args{
					{
						Name:        "package",
						Description: "Package name",
					},
				},
				Description: "Prepare a package for patchin",
				Flags:       globalFlags(nil),
				Execute:     inst.run,
			},
			{
				Name:        "outdated",
				Description: "Display latest versions of outdated dependencies",
				Flags:       globalFlags(nil),
				Execute:     inst.run,
			},
			{
				Name:        "pm",
				Description: "Additional package management utilities",
				Args: tree.Args{
					{
						Name:        "subcommand",
						Description: "Package command",
					},
				},
				Flags:   globalFlags(nil),
				Execute: inst.run,
			},
		},
		Execute: inst.run,
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

// func (c *Command) run(ctx context.Context, dirname, script string) error {
// 	return shell.New(ctx, c.l, "bun", "run", script).
// 		Dir(dirname).
// 		Run()
// }

func (c *Command) run(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "bun").
		Args(r.Args()...).
		Args(r.Flags()...).
		Run()
}

func (c *Command) runWorkspace(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "bun").
		Dir(r.Args().At(1)).Args(r.Args().From(2)...).
		Args(r.Flags()...).
		Run()
}

//nolint:forcetypeassert
func (c *Command) scripts(ctx context.Context, dirname string) []string {
	return c.cache.Get("scripts-"+dirname, func() any {
		var ret []string

		f, err := npm.LoadPackageJSON(path.Join(dirname, "package.json"))
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
		var (
			ret   []string
			roots []string
		)

		{
			value, err := npm.LoadPackageJSON("package.json")
			if err != nil {
				c.l.Debug("failed to load package.json", err.Error())
				return []string{}
			}

			for _, s := range value.Workspaces {
				roots = append(roots, strings.TrimSuffix(s, "/*"))
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
