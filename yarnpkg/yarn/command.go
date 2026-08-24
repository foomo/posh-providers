package yarn

import (
	"context"
	_ "embed"
	"os"
	"path"
	"strings"

	"github.com/cloudrecipes/packagejson"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"golang.org/x/sync/errgroup"
)

//go:embed SKILL.md
var skill string

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

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

func NewCommand(l log.Logger, c cache.Cache, opts ...CommandOption) *Command {
	inst := &Command{
		l:     l.Named("yarn"),
		name:  "yarn",
		cache: c.Get("yarn"),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run yarn commands",
		Execute:     inst.execute,
		Nodes: tree.Nodes{
			&tree.Node{
				Name:        "install",
				Description: "Install dependencies",
				Args:        tree.Args{inst.pathArg()},
				Execute:     inst.install,
			},
			&tree.Node{
				Name:        "run",
				Description: "Run script",
				Args: tree.Args{
					&tree.Arg{
						Name:        "path",
						Description: "Directory containing the package.json; required here, the script is read from the next argument",
						Optional:    true,
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.paths(ctx))
						},
					},
					&tree.Arg{
						Name:        "script",
						Description: "package.json script to run; completed from the chosen path",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.scripts(ctx, r.Args().At(1)))
						},
					},
				},
				Execute: inst.run,
			},
			&tree.Node{
				Name:        "run-all",
				Description: "Run script in every discovered package, excluding the root",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Default().Int("parallel", 0, "number of parallel processes")
					return nil
				},
				Args: tree.Args{
					&tree.Arg{
						Name:        "script",
						Description: "package.json script to run in every discovered package; not completed",
					},
				},
				Execute: inst.runAll,
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog lists
// three subcommands and cannot show that the root forwards everything else to
// the yarn binary, so `add`, `upgrade` and `publish` are reachable and unlisted.
// Nor can it show that `run`'s `[path]` is marked optional but read by index, so
// omitting it takes the script as the directory; that `run-all` skips the root
// package; that discovery ignores any directory whose name merely contains
// "dist"; or that the subcommands drop flags typed before `--`.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names running a script across packages, because that is the form with both the
// positional-argument trap and the concurrent fan-out worth loading the file for.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when managing this project's JavaScript or TypeScript dependencies with " +
			"yarn - installing dependencies in the root or one package, running a package.json " +
			"script in a chosen directory, running the same script across every nested package, " +
			"or reaching an upstream verb like add, upgrade or publish through the passthrough.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "yarn").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) run(ctx context.Context, r *readline.Readline) error {
	dir, script := r.Args().At(1), r.Args().At(2)
	c.l.Infof("Running script %q in %q", script, dir)

	return shell.New(ctx, c.l, "yarn", "run", script).
		Args(r.AdditionalArgs()...).
		Dir(dir).
		Run()
}

func (c *Command) runAll(ctx context.Context, r *readline.Readline) error {
	script := r.Args().At(1)
	ctx, wg := c.wg(ctx, r)
	c.l.Infof("Running script %q in...", script)

	for _, dir := range c.paths(ctx) {
		if dir != "." {
			wg.Go(func() error {
				c.l.Info("└ " + dir)

				return shell.New(ctx, c.l, "yarn", "run", script).
					Args(r.AdditionalArgs()...).
					Dir(dir).
					Run()
			})
		}
	}

	return wg.Wait()
}

func (c *Command) install(ctx context.Context, r *readline.Readline) error {
	dir := "."
	if r.Args().LenGt(1) {
		dir = r.Args().At(1)
	}

	c.l.Infof("Running install in %q", dir)

	return shell.New(ctx, c.l, "yarn", "install").
		Args(r.AdditionalArgs()...).
		Dir(dir).
		Run()
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		if value, err := files.Find(ctx, ".", "package.json", files.FindWithIgnore(`^\.`, "dist", "node_modules")); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else {
			for i, s := range value {
				value[i] = path.Dir(s)
			}

			return value
		}
	}).([]string)
}

//nolint:forcetypeassert
func (c *Command) scripts(ctx context.Context, filename string) []string {
	return c.cache.Get("scripts-"+strings.ReplaceAll(filename, "/", "-"), func() any {
		payload, err := os.ReadFile(path.Join(filename, "package.json"))
		if err != nil {
			return []string{}
		}

		ret := make([]string, 0)

		if value, err := packagejson.Parse(payload); err != nil {
			return nil
		} else {
			for name := range value.Scripts {
				ret = append(ret, name)
			}
		}

		return ret
	}).([]string)
}

func (c *Command) pathArg() *tree.Arg {
	return &tree.Arg{
		Name:        "path",
		Description: "Directory containing the package.json; the project root if omitted",
		Optional:    true,
		Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
			return suggests.List(c.paths(ctx))
		},
	}
}

func (c *Command) wg(ctx context.Context, r *readline.Readline) (context.Context, *errgroup.Group) {
	wg, ctx := errgroup.WithContext(ctx)
	if value, err := r.FlagSets().Default().GetInt("parallel"); err == nil && value != 0 {
		wg.SetLimit(value)
	} else {
		wg.SetLimit(1)
	}

	return ctx, wg
}
