package gherkinlint

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
	l               log.Logger
	name            string
	cache           cache.Namespace
	execGherkinLint exec.CommandProvider
	commandTree     tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) options.Option[*Command] {
	return func(o *Command) {
		o.name = v
	}
}

func CommandWithExecGherkinLint(v exec.CommandProvider) options.Option[*Command] {
	return func(o *Command) {
		o.execGherkinLint = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, c cache.Cache, opts ...options.Option[*Command]) *Command {
	inst := &Command{
		l:     l.Named("gherkin-lint"),
		name:  "gherkin-lint",
		cache: c.Get("gherkin-lint"),
		execGherkinLint: func(ctx context.Context, args ...string) *exec.Command {
			return exec.NewCommand(ctx, "gherkin-lint", args...)
		},
	}

	options.Apply(inst, opts...)

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run gherkin-lint",
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

func (c *Command) Lint(ctx context.Context, _ bool) error {
	return c.run(ctx, c.paths(ctx))
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	paths := r.Args().From(0)
	if len(paths) == 0 {
		paths = c.paths(ctx)
	}

	return c.run(ctx, paths)
}

func (c *Command) run(ctx context.Context, paths []string) error {
	c.l.Info("Running gherkin-lint ...")

	for _, dir := range paths {
		c.l.Info("└ " + dir)

		if err := c.execGherkinLint(ctx, dir).Run(); err != nil {
			return err
		}
	}

	return nil
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		matches, err := files.Find(ctx, ".", "wdio.conf.ts", files.FindWithIgnore(`^(node_modules|\.\w*)$`))
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
