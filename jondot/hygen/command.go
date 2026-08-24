package hygen

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

//go:embed SKILL.md
var skill string

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

type (
	Command struct {
		l           log.Logger
		cfg         Config
		name        string
		cache       cache.Namespace
		configKey   string
		commandTree tree.Root
	}
	Option func(*Command) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) Option {
	return func(o *Command) error {
		o.name = v
		return nil
	}
}

func WithConfigKey(v string) Option {
	return func(o *Command) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, opts ...Option) (*Command, error) {
	inst := &Command{
		l:         l.Named("hygen"),
		name:      "hygen",
		configKey: "hygen",
		cache:     cache.Get("hygen"),
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(inst); err != nil {
				return nil, err
			}
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Scaffold files from a hygen template",
		Args: tree.Args{
			{
				Name:        "path",
				Description: "Unused: the root has no execute, so only the template subcommand runs",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.paths(ctx))
				},
			},
		},
		Nodes: tree.Nodes{
			{
				Name:        "template",
				Description: "Render a template directory from the configured template path",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.paths(ctx))
				},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Default().Bool("dry", false, "Perform a dry run. Files will be generated but not saved")
					return nil
				},
				Args: tree.Args{
					{
						Name:        "path",
						Description: "Target path passed through to hygen; not completed",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return nil
						},
					},
				},
				Execute: inst.execute,
			},
		},
	})

	return inst, nil
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
		return errors.New("missing [template] argument")
	case r.Args().LenIs(1):
		return errors.New("missing [path] argument")
	case r.Args().LenGt(2):
		return errors.New("too many arguments")
	}

	if info, err := os.Stat(filepath.Join(c.cfg.TemplatePath, r.Args().At(0))); err != nil || !info.IsDir() {
		return errors.Errorf("invalid [TEMPLATE] parameter: %s", r.Args().At(0))
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
// cannot show that only the `template` leaf is runnable, that the leaf's own
// name is forwarded to hygen as an argument, that validation stats that literal
// word rather than the named template, or that HYGEN_TMPLS is the parent of the
// configured template path.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// leads with scaffolding a new component from a template, since that is the
// request an agent will actually receive - the word "hygen" rarely appears in it.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when scaffolding new files from one of this project's own code " +
			"templates - generating a component, service, package or module skeleton, asking " +
			"which scaffolds exist, or previewing what a template would write before it " +
			"writes it. Also for hygen templates and the HYGEN_TMPLS directory.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "hygen", "scaffold").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Env(fmt.Sprintf("HYGEN_TMPLS=%s", path.Dir(c.cfg.TemplatePath))).
		Run()
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		files, err := os.ReadDir(c.cfg.TemplatePath)
		if err != nil {
			c.l.Debug("failed to read template dir:", err.Error())
			return []string{}
		}

		ret := make([]string, 0, len(files))
		for _, value := range files {
			if value.IsDir() {
				ret = append(ret, value.Name())
			}
		}

		return ret
	}).([]string)
}
