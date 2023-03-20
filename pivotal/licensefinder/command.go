package licensefinder

import (
	"context"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/pkg/errors"
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

func NewCommand(l log.Logger, cache cache.Cache, opts ...CommandOption) *Command {
	inst := &Command{
		l:         l.Named("licensefinder"),
		name:      "licensefinder",
		configKey: "licensefinder",
		cache:     cache.Get("licensefinder"),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	addNode := tree.Node{
		Name:        "add",
		Args:        tree.Args{{Name: "name"}},
		Description: "Add entry",
		Execute:     inst.execute,
	}
	listNode := tree.Node{
		Name:        "list",
		Description: "List entry",
		Execute:     inst.execute,
	}
	removeNode := tree.Node{
		Name:        "remove",
		Args:        tree.Args{{Name: "name"}},
		Description: "Remove entry",
		Execute:     inst.execute,
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run license finder",
		Execute:     inst.execute,
		Nodes: tree.Nodes{
			{
				Name:        "restricted_licenses",
				Description: "Manage restricted licenses",
				Nodes: tree.Nodes{
					&addNode,
					&listNode,
					&removeNode,
				},
			},
			{
				Name:        "ignored_dependencies",
				Description: "Manage ignored dependencies",
				Nodes: tree.Nodes{
					&addNode,
					&listNode,
					&removeNode,
				},
			},
			{
				Name:        "permitted_licenses",
				Description: "Manage permitted licenses",
				Nodes: tree.Nodes{
					&addNode,
					&listNode,
					&removeNode,
				},
			},
			{
				Name:        "approvals",
				Description: "Manage approvals",
				Nodes: tree.Nodes{
					&addNode,
					&removeNode,
				},
			},
			{
				Name:        "licenses",
				Description: "Manage licenses",
				Nodes: tree.Nodes{
					&addNode,
					&removeNode,
				},
			},
		},
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
	if _, err := exec.LookPath("license_finder"); err != nil {
		c.l.Print()
		return errors.New(`
Please ensure you have the license_finder installed!

- Install binary:

$ brew update
$ brew install licensefinder
        `)
	}
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

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	var paths []string
	args := []string{
		"--log-directory=" + c.cfg.LogPath,
		"--decisions_file=" + c.cfg.DecisionsPath,
	}
	if r.Args().LenIs(0) {
		paths = append(paths, c.paths(ctx, "go.sum")...)
		paths = append(paths, c.paths(ctx, "yarn.lock")...)
		args = append(args, "--aggregate_paths="+strings.Join(paths, " "))
	}
	return shell.New(ctx, c.l, "license_finder").
		Args(args...).
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context, filename string) []string {
	return c.cache.Get("paths-"+filename, func() any {
		if value, err := files.Find(ctx, ".", filename); err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return nil
		} else {
			for i, s := range value {
				value[i] = path.Dir(s)
			}
			return value
		}
	}).([]string)
}
