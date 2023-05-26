package licensefinder

import (
	"context"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/foomo/posh/pkg/cache"
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
		Description: "Name of the license",
	}
	addFlags := func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
		fs.Default().String("who", "", "Who approved")
		fs.Default().String("why", "", "Reason to approve")
		return nil
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "List unapproved dependencies",
		Execute:     inst.actionItems,
		Nodes: tree.Nodes{
			{
				Name:        "report",
				Description: "Print a report of the project's dependencies",
				Execute:     inst.report,
			},
			{
				Name:        "add",
				Description: "Add licenses or dependencies",
				Nodes: tree.Nodes{
					{
						Name:        "permitted",
						Description: "Add permitted licenses",
						Args:        tree.Args{nameArg},
						Flags:       addFlags,
						Execute:     inst.addPermitted,
					},
					{
						Name:        "ignored",
						Description: "Add ignored dependencies",
						Args:        tree.Args{nameArg},
						Flags:       addFlags,
						Execute:     inst.addIgnored,
					},
				},
			},
			{
				Name:        "list",
				Description: "List licenses or dependencies",
				Nodes: tree.Nodes{
					{
						Name:        "permitted",
						Description: "Add permitted licenses",
						Execute:     inst.listPermitted,
					},
					{
						Name:        "ignored",
						Description: "List ignored dependencies",
						Args:        tree.Args{nameArg},
						Execute:     inst.listIgnored,
					},
				},
			},
			{
				Name:        "remove",
				Description: "Remove licenses or dependencies",
				Nodes: tree.Nodes{
					{
						Name:        "permitted",
						Description: "Add permitted licenses",
						Args:        tree.Args{nameArg},
						Execute:     inst.removePermitted,
					},
					{
						Name:        "ignored",
						Description: "Remove ignored dependencies",
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
	return c.execute(ctx, r, "action_items", c.aggregatePaths(ctx))
}

func (c *Command) report(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r, "report", c.aggregatePaths(ctx))
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

func (c *Command) aggregatePaths(ctx context.Context) string {
	var paths []string
	paths = append(paths, c.paths(ctx, "go.sum")...)
	paths = append(paths, c.paths(ctx, "yarn.lock")...)
	paths = lo.Uniq(paths)
	sort.Strings(paths)
	c.l.Info("Aggregating liceses from:")
	for _, value := range paths {
		c.l.Info("└  " + value)
	}
	return "--aggregate_paths=" + strings.Join(paths, " ")
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
