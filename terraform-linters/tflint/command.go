package tflint

import (
	"context"
	"path"

	"github.com/foomo/go/options"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/exec"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
)

type Command struct {
	l           log.Logger
	name        string
	cache       cache.Namespace
	execTflint  exec.CommandProvider
	commandTree tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) options.Option[*Command] {
	return func(o *Command) {
		o.name = v
	}
}

func CommandWithExecTflint(v exec.CommandProvider) options.Option[*Command] {
	return func(o *Command) {
		o.execTflint = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, c cache.Cache, opts ...options.Option[*Command]) *Command {
	inst := &Command{
		l:     l.Named("tflint"),
		name:  "tflint",
		cache: c.Get("tflint"),
		execTflint: func(ctx context.Context, args ...string) *exec.Command {
			return exec.NewCommand(ctx, "tflint", args...)
		},
	}

	options.Apply(inst, opts...)

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run tflint",
		Args: tree.Args{
			{
				Name:     "path",
				Optional: true,
				Repeat:   true,
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.paths(ctx))
				},
			},
		},
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Default().Bool("fix", false, "run quick fix")
			return nil
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

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

func (c *Command) Lint(ctx context.Context, fix bool) error {
	return c.run(ctx, fix, c.paths(ctx))
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	fix, err := r.FlagSets().Default().GetBool("fix")
	if err != nil {
		return err
	}

	paths := r.Args().From(0)
	if len(paths) == 0 {
		paths = c.paths(ctx)
	}

	return c.run(ctx, fix, paths)
}

func (c *Command) run(ctx context.Context, fix bool, paths []string) error {
	c.l.Info("Running tflint ...")

	var args []string
	if fix {
		args = append(args, "--fix")
	}

	for _, dir := range paths {
		c.l.Info("└ " + dir)

		if err := c.execTflint(ctx, args...).Dir(dir).Run(); err != nil {
			return err
		}
	}

	return nil
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		matches, err := files.Find(ctx, ".", "main.tf", files.FindWithIgnore(`^(node_modules|\.\w*)$`))
		if err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		}

		out := make([]string, 0, len(matches))
		for _, m := range matches {
			out = append(out, path.Dir(m))
		}

		return out
	}).([]string)
}
