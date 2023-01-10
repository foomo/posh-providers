package zeus

import (
	"context"
	"os"
	"path"
	"strings"

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
		l:     l.Named("zeus"),
		name:  "zeus",
		cache: cache.Get("zeus"),
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
	switch {
	case r.Args().LenLte(1):
		return c.completePaths(ctx)
	}
	return nil
}

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	switch {
	case r.Args().LenLt(1):
		return errors.New("missing [path] parameter")
	}
	dir := r.Args().At(0)
	if info, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		if !strings.HasSuffix(dir, "/zeus") {
			return errors.Errorf("invalid [path] parameter: %s", dir)
		}
	} else if err != nil || !info.IsDir() {
		return errors.Errorf("invalid [path] parameter: %s", dir)
	}
	return nil
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	dir, args := r.Args().Shift()
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		c.cache.Delete("")
		c.l.Info("bootstrapping a new zeus:", dir)
		return shell.New(ctx, c.l, "zeus", "bootstrap").
			Args(args...).
			Args(r.PassThroughArgs()...).
			Args(r.AdditionalArgs()...).
			Dir(path.Join(dir, "..")).
			Run()
	} else {
		return shell.New(ctx, c.l, "zeus", "-C", path.Dir(dir)).
			Args(args...).
			Args(r.PassThroughArgs()...).
			Args(r.AdditionalArgs()...).
			Run()
	}
}

func (c *Command) Help() string {
	return `Find and run zeus at the given path.

If the given path doesn't exist, it will bootstrap a new zeus installation.

Usage:
  zeus [path] <args>...

Examples:
  gomod tidy ./path
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
		if value, err := files.Find(ctx, "zeus"); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return nil
		} else {
			return value
		}
	}).([]string)
}
