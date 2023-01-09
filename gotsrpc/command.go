package gotsrpc

import (
	"context"
	"os"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

type Command struct {
	l     log.Logger
	name  string
	cache cache.Namespace
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache) *Command {
	return &Command{
		l:     l,
		name:  "gotsrpc",
		cache: cache.Get("gotsrpc"),
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.name
}

func (c *Command) Description() string {
	return "run zeus on target"
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline, d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "-debug"},
		{Text: "-skipgotsrpc"},
	}
}

func (c *Command) CompleteArguments(ctx context.Context, r *readline.Readline, d prompt.Document) []prompt.Suggest {
	switch {
	case r.Args().LenLte(1):
		return c.completePaths(ctx)
	}
	return nil
}

// Validate ...
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

// Execute ...
func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(0) {
		paths = []string{r.Args().At(0)}
	} else {
		paths = c.paths(ctx)
	}
	for _, value := range paths {
		c.l.Info("gotsrpc:", value)
		if out, err := shell.New(ctx, c.l, "gotsrpc").
			Args(r.Flags()...).
			Args(value).
			Args(r.AdditionalArgs()...).
			Output(); err != nil {
			return errors.Wrap(err, string(out))
		}
	}
	return nil
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

func (c *Command) completePaths(ctx context.Context) []prompt.Suggest {
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
