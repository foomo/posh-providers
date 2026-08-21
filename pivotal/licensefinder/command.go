package licensefinder

import (
	"context"
	_ "embed"
	"os/exec"
	"path"
	"sort"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

//go:embed SKILL.md
var skill string

type (
	Command struct {
		l           log.Logger
		cfg         Config
		name        string
		cache       cache.Namespace
		configKey   string
		commandTree tree.Root
	}
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Option
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

func CommandWithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("licensefinder"),
		name:      "licensefinder",
		configKey: "licenseFinder",
		cache:     cache.Get("licensefinder"),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	nameArg := &tree.Arg{
		Name:        "name",
		Description: "License or dependency name, exactly as license_finder reports it",
	}
	addFlags := func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
		fs.Default().String("who", "", "Person recording the decision, stored in the decisions file")
		fs.Default().String("why", "", "Justification, stored in the decisions file")

		return nil
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "List dependencies whose licenses are not yet approved",
		Execute:     inst.actionItems,
		Nodes: tree.Nodes{
			{
				Name:        "report",
				Description: "Print a license report for every discovered project",
				Execute:     inst.report,
			},
			{
				Name:        "add",
				Description: "Record an approval decision in the decisions file",
				Nodes: tree.Nodes{
					{
						Name:        "permitted",
						Description: "Permit a license for every dependency that uses it",
						Args:        tree.Args{nameArg},
						Flags:       addFlags,
						Execute:     inst.addPermitted,
					},
					{
						Name:        "ignored",
						Description: "Exclude a dependency from license checks entirely",
						Args:        tree.Args{nameArg},
						Flags:       addFlags,
						Execute:     inst.addIgnored,
					},
				},
			},
			{
				Name:        "list",
				Description: "Show recorded approval decisions",
				Nodes: tree.Nodes{
					{
						Name:        "permitted",
						Description: "List the permitted licenses",
						Execute:     inst.listPermitted,
					},
					{
						Name:        "ignored",
						Description: "List the ignored dependencies",
						Execute:     inst.listIgnored,
					},
				},
			},
			{
				Name:        "remove",
				Description: "Delete a recorded approval decision",
				Nodes: tree.Nodes{
					{
						Name:        "permitted",
						Description: "Remove a permitted license",
						Args:        tree.Args{nameArg},
						Execute:     inst.removePermitted,
					},
					{
						Name:        "ignored",
						Description: "Stop ignoring a dependency",
						Args:        tree.Args{nameArg},
						Execute:     inst.removeIgnored,
					},
				},
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
	if _, err := exec.LookPath("license_finder"); err != nil {
		c.l.Print()

		return errors.New(`
Please ensure you have the license_finder installed!

- Install binary:

$ brew update
$ brew install licensefinder
        `)
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

// Skill implements the optional command.Skiller interface. The catalog cannot
// show that the add/remove verbs rewrite a version-controlled decisions file and
// are policy changes rather than build steps, that a non-zero exit is the
// intended signal while a zero exit can mean nothing was scanned, or that
// every discovered source path is aggregated into one combined report.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) addPermitted(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r, append([]string{"permitted_licenses", "add"}, r.Args().From(2)...)...)
}

func (c *Command) listPermitted(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r, "permitted_licenses", "list")
}

func (c *Command) removePermitted(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r, append([]string{"permitted_licenses", "remove"}, r.Args().From(2)...)...)
}

func (c *Command) addIgnored(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r, append([]string{"ignored_dependencies", "add"}, r.Args().From(2)...)...)
}

func (c *Command) listIgnored(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r, "ignored_dependencies", "list")
}

func (c *Command) removeIgnored(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r, append([]string{"ignored_dependencies", "remove"}, r.Args().From(2)...)...)
}

func (c *Command) actionItems(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r, append([]string{"action_items"}, c.aggregatePaths(ctx)...)...)
}

func (c *Command) report(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r, append([]string{"report"}, c.aggregatePaths(ctx)...)...)
}

func (c *Command) execute(ctx context.Context, r *readline.Readline, args ...string) error {
	fs := r.FlagSets().Default()

	return shell.New(ctx, c.l, "license_finder").
		Args(args...).
		Args(
			"--log-directory="+c.cfg.LogPath,
			"--decisions-file="+c.cfg.DecisionsPath,
		).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

// aggregatePaths returns the --aggregate_paths flag and one quoted argument per
// discovered source directory.
func (c *Command) aggregatePaths(ctx context.Context) []string {
	var paths []string
	for _, file := range c.cfg.Sources {
		paths = append(paths, c.paths(ctx, file)...)
	}

	paths = lo.Uniq(paths)
	sort.Strings(paths)
	c.l.Info("Aggregating liceses from:")

	for _, value := range paths {
		c.l.Info("└  " + value)
	}

	if len(paths) == 0 {
		return nil
	}

	ret := []string{"--aggregate_paths"}
	for _, value := range paths {
		ret = append(ret, "'"+value+"'")
	}

	return ret
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context, filename string) []string {
	return c.cache.Get("paths-"+filename, func() any {
		if value, err := files.Find(ctx, ".", filename, files.FindWithIgnore(`^\.`, "vendor", "node_modules")); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		} else {
			for i, s := range value {
				value[i] = path.Dir(s)
			}

			return value
		}
	}).([]string)
}
