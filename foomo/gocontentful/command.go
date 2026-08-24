package gocontentful

import (
	"context"
	_ "embed"
	"os"
	"path"
	"strings"

	"github.com/foomo/posh-providers/onepassword"
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
	"gopkg.in/yaml.v3"
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
		op          *onepassword.OnePassword
		name        string
		cache       cache.Namespace
		commandTree tree.Root
	}
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, op *onepassword.OnePassword, opts ...CommandOption) *Command {
	inst := &Command{
		l:     l.Named("gocontentful"),
		name:  "gocontentful",
		cache: cache.Get("gocontentful"),
		op:    op,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Generate Contentful client code from gocontentful.yaml files",
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Default().Bool("debug", false, "show debug output")
			return nil
		},
		Args: tree.Args{
			{
				Name:        "path",
				Description: "Path to a gocontentful.yaml; every one found in the project if omitted",
				Optional:    true,
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
// cannot show that each run deletes gocontentful*.go before regenerating, that
// an omitted path does so for every config in the project, that a 1Password
// secret is resolved and sent to the Contentful API, or that `Config` describes
// the discovered yaml file rather than a posh config key.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The triggers name
// the generated artefacts, since the usual way in is a compile error or a stale
// type after a content model change rather than a request naming this command.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when this project's generated Contentful client code needs " +
			"regenerating - after a content type or content model change, or when the " +
			"gocontentful*.go files are stale, missing or fail to compile. Also for " +
			"gocontentful.yaml files and Contentful space, environment or content type " +
			"configuration.",
	}
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

	c.l.Info("Running gocontentful...")

	for _, value := range paths {
		c.l.Info("└  " + value)

		var cfg Config
		if out, err := c.op.RenderFile(ctx, value); err != nil {
			return errors.Wrap(err, "failed to render gocontentful.yaml")
		} else if err := yaml.Unmarshal(out, &cfg); err != nil {
			return errors.Wrap(err, "failed to unmarshal gocontentful.yaml")
		}

		dir := path.Dir(value)
		if out, err := shell.New(ctx, c.l, "rm", "-f", dir+"/gocontentful*.go").Output(); err != nil {
			return errors.Wrap(err, string(out))
		}

		// "master" is the built-in default environment name of contentful
		// fallback to it if the configuration does not ask specifically for an environment
		if cfg.Environment == "" {
			cfg.Environment = "master"
		}

		if err := shell.New(ctx, c.l, "gocontentful",
			"-spaceid", cfg.SpaceID, "-cmakey", cfg.CMAKey, "-environment", cfg.Environment,
			"-region", cfg.Region,
			"-contenttypes", strings.Join(cfg.ContentTypes, ","), dir).
			Args(r.AdditionalArgs()...).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context) []string {
	return c.cache.Get("paths", func() any {
		if value, err := files.Find(ctx, ".", "gocontentful.yaml", files.FindWithIgnore(`^\.`, "node_modules")); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else {
			return value
		}
	}).([]string)
}
