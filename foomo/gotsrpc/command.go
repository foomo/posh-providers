package gotsrpc

import (
	"context"
	"os"
	"strings"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
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
		l:     l.Named("gotsrpc"),
		cache: cache.Get("gotsrpc"),
	}
	inst.commandTree = &tree.Root{
		Name:        "gotsrpc",
		Description: "run gotsrpc",
		Node: &tree.Node{
			Flags: func(fs *readline.FlagSet) {
				fs.Bool("debug", false, "show debug output")
			},
			Args: tree.Args{
				{
					Name:    "path",
					Suggest: inst.completePaths,
				},
			},
			Execute: inst.execute,
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

func (c *Command) Complete(ctx context.Context, r *readline.Readline, d prompt.Document) []prompt.Suggest {
	return c.commandTree.RunCompletion(ctx, r)
}

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	switch {
	case r.Args().LenIs(0):
		return nil
	case r.Args().LenGt(1):
		return errors.New("too many arguments")
	}

	if info, err := os.Stat(r.Args().At(0)); err != nil || info.IsDir() {
		return errors.New("invalid [path] parameter")
	}

	return nil
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.RunExecution(ctx, r)
}

func (c *Command) Help() string {
	return `Generate gotsrpc files.

Usage:
  gotsrpc <path> <options>

Available options:
  debug
  skipgotsrpc

Examples:
  gotsrpc ./path/gotsrpc.yml
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(0) {
		paths = []string{r.Args().At(0)}
	} else {
		paths = c.paths(ctx)
	}
	flags := make([]string, len(r.Flags()))
	for i, flag := range r.Flags() {
		flags[i] = strings.ReplaceAll(flag, "--", "-")
	}

	for _, value := range paths {
		c.l.Info("gotsrpc:", value)
		if out, err := shell.New(ctx, c.l, "gotsrpc").
			Args(flags...).
			Args(value).
			Args(r.AdditionalArgs()...).
			Output(); err != nil {
			return errors.Wrap(err, string(out))
		}
	}
	return nil
}

func (c *Command) completePaths(ctx context.Context, t *tree.Root, r *readline.Readline) []prompt.Suggest {
	return suggests.List(c.paths(ctx))
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		if value, err := files.Find(ctx, "gotsrpc.yml"); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return nil
		} else {
			return value
		}
	}).([]string)
}
