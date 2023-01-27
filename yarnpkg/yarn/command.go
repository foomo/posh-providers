package yarn

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/cloudrecipes/packagejson"
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
		name        string
		cache       cache.Namespace
		commandTree *tree.Root
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

	inst.commandTree = &tree.Root{
		Name:        inst.name,
		Description: "run yarn commands",
		Nodes: tree.Nodes{
			{
				Name:        "install",
				Description: "install dependencies",
				Args:        tree.Args{inst.pathArg()},
				Execute:     inst.install,
			},
			{
				Name:        "run",
				Description: "run script",
				Args: tree.Args{
					inst.pathArg(),
					{
						Name:     "script",
						Repeat:   false,
						Optional: false,
						Suggest: func(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.scripts(ctx, r.Args().At(1)))
						},
					},
				},
				Execute: inst.run,
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

func (c *Command) Help() string {
	return `Run yarn commands.

Usage:
  yarn [cmd]

Available Commands:
  install <path>        install dependencies
  run [path] [script]   run script
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) run(ctx context.Context, r *readline.Readline) error {
	dir, script := r.Args().At(0), r.Args().At(1)
	return shell.New(ctx, c.l, "yarn", "run", script).
		Args(r.PassThroughFlags()...).
		Args(r.AdditionalArgs()...).
		Dir(dir).
		Run()
}

func (c *Command) install(ctx context.Context, r *readline.Readline) error {
	dir := "."
	if r.Args().LenGt(1) {
		dir = r.Args().At(1)
	}
	return shell.New(ctx, c.l, "yarn", "install").
		Args(r.PassThroughFlags()...).
		Args(r.AdditionalArgs()...).
		Dir(dir).
		Run()
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		if value, err := files.Find(ctx, ".", "package.json", files.FindWithIgnore("node_modules")); err != nil {
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
			return nil
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
		Name:     "path",
		Optional: true,
		Suggest: func(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
			return suggests.List(c.paths(ctx))
		},
	}
}
