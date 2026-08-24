package lint

import (
	"context"
	_ "embed"
	"slices"
	"strings"

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

//go:embed SKILL.md
var skill string

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

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
		Description: "Run every registered linter, or the named ones",
		Args: tree.Args{
			{
				Name:        "name",
				Description: "Linter name; every registered linter if omitted",
				Optional:    true,
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.linterNames())
				},
			},
		},
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Default().Bool("fix", false, "let each linter rewrite files to fix findings, where it supports it")
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The tree shows one
// argument and one flag but not which linters are registered, that `--fix` is
// forwarded to each of them and may rewrite files, or that they all run
// concurrently under a shared context where the first failure cancels the rest.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the intents that should reach an aggregator - linting everything at once,
// auto-fixing findings - rather than any single linter, because which linters
// exist is a per-project fact this text cannot know.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when linting this project - running every configured linter at once, " +
			"linting only named ones, or auto-fixing findings with --fix. Also when a lint or " +
			"static-analysis check fails in CI and needs reproducing locally, or when asking " +
			"which linters this project even has.",
	}
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

		// Warn per name that matched nothing. The filter above only ever errored
		// when *every* name missed, so `lint go bogus` ran `go` and never
		// mentioned `bogus` - which reads as "both linters passed".
		for _, name := range names {
			if !slices.ContainsFunc(linters, func(lt Linter) bool { return lt.Name() == name }) {
				c.l.Warn("unknown linter:", name)
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
