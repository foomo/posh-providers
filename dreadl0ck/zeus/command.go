package zeus

import (
	"context"
	_ "embed"
	"os"
	"path"
	"strings"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

//go:embed SKILL.md
var skill string

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
	return "Run zeus in the given directory, or bootstrap a new one"
}

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command. The command is not tree based, so
// the CommandInfo is built by hand.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return command.CommandInfo{
		FullPath:    c.name,
		Description: c.Description(),
		Arguments: []command.ArgInfo{
			{
				Name:        "path",
				Description: "Path of the zeus directory; bootstraps a new one if it does not exist",
			},
			{
				Name:        "args",
				Description: "Arguments forwarded verbatim to the zeus binary",
				Optional:    true,
				Repeat:      true,
			},
		},
	}
}

// Skill implements the optional command.Skiller interface. The command has no
// subcommand tree at all, so the catalog can only show two placeholders: it
// cannot show that the forwarded arguments are the real surface, that a
// non-existent path bootstraps rather than errors, or that the path is the zeus
// directory whose parent zeus is actually run in.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
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
		c.cache.Delete()
		c.l.Info("bootstrapping a new zeus:", dir)

		return shell.New(ctx, c.l, "zeus", "bootstrap").
			Args(args...).
			Args(r.AdditionalArgs()...).
			Dir(path.Join(dir, "..")).
			Run()
	} else {
		return shell.New(ctx, c.l, "zeus", "-C", path.Dir(dir)).
			Args(args...).
			Args(r.AdditionalArgs()...).
			Run()
	}
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return `Find and run zeus at the given path.

The path is the zeus directory itself; zeus is run in its parent directory.
If the given path doesn't exist, it will bootstrap a new zeus installation.

Usage:
  zeus [path] <args>...

Examples:
  zeus ./svc/zeus build
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) completePaths(ctx context.Context) []goprompt.Suggest {
	return suggests.List(c.paths(ctx))
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		if value, err := files.Find(ctx, ".", "zeus", files.FindWithIgnore(`^\.`, "node_modules"), files.FindWithIsDir(true)); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else {
			return value
		}
	}).([]string)
}
