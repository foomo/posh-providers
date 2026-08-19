package gotsrpc

import (
	"context"
	_ "embed"
	"os"
	"strings"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
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
	l           log.Logger
	cache       cache.Namespace
	commandTree tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache) *Command {
	inst := &Command{
		l:     l.Named("gotsrpc"),
		cache: cache.Get("gotsrpc"),
	}
	inst.commandTree = tree.New(&tree.Node{
		Name:        "gotsrpc",
		Description: "Generate gotsrpc code from gotsrpc.yml files",
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Default().Bool("debug", false, "show debug output")
			return nil
		},
		Args: tree.Args{
			{
				Name:        "path",
				Description: "Path to a gotsrpc.yml file; every one found in the project if omitted",
				Optional:    true,
				Suggest:     inst.completePaths,
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

// Skill implements the optional command.Skiller interface. The rendered tree
// cannot show that an omitted path regenerates from every gotsrpc.yml in the
// project, that doing so overwrites generated sources named only in those
// files, or that forwarded flags have every "--" rewritten to "-".
func (c *Command) Skill(ctx context.Context) string {
	return skill
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

func (c *Command) completePaths(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	return suggests.List(c.paths(ctx))
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		if value, err := files.Find(ctx, ".", "gotsrpc.yml"); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return nil
		} else {
			return value
		}
	}).([]string)
}
