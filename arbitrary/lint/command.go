package lint

import (
	"context"
	"slices"

	"github.com/foomo/go/options"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Command struct {
	l           log.Logger
	name        string
	commands    command.Commands
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

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, commands command.Commands, opts ...options.Option[*Command]) *Command {
	inst := &Command{
		l:        l.Named("lint"),
		name:     "lint",
		commands: commands,
	}

	options.Apply(inst, opts...)

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Lint your code",
		Args: tree.Args{
			{
				Name:     "name",
				Optional: true,
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.linterNames())
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

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()
	linters := c.linters()

	fix, err := fs.GetBool("fix")
	if err != nil {
		return err
	}

	if r.Args().LenGt(0) {
		names := r.Args().From(0)
		linters = []Linter{}

		for _, lt := range c.linters() {
			if slices.Contains(names, lt.Name()) {
				linters = append(linters, lt)
			}
		}

		if len(linters) == 0 {
			return errors.Errorf("unknown linter: %s", names)
		}
	}

	wg, ctx := errgroup.WithContext(ctx)

	for _, lt := range linters {
		c.l.Info("Linting with " + lt.Name() + " ...")

		wg.Go(func() error {
			return lt.Lint(ctx, fix)
		})
	}

	return wg.Wait()
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) linters() []Linter {
	var ret []Linter

	for _, value := range c.commands.List() {
		if v, ok := value.(Linter); ok {
			ret = append(ret, v)
		}
	}

	return ret
}

func (c *Command) linterNames() []string {
	var ret []string
	for _, linter := range c.linters() {
		ret = append(ret, linter.Name())
	}

	slices.Sort(ret)

	return ret
}
