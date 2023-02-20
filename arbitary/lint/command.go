package lint

import (
	"context"
	"path"
	"strings"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

type Linter string

type (
	Command struct {
		l           log.Logger
		name        string
		cache       cache.Namespace
		commandTree *tree.Root
	}
	CommandOption func(command *Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

func CommandWithGo() CommandOption {
	return func(o *Command) {
		o.commandTree.Nodes = append(o.commandTree.Nodes, &tree.Node{
			Name:        "go",
			Description: "run golangci-lint",
			Args:        tree.Args{o.pathArg("go.mod")},
			Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
				fs.Bool("fix", false, "run quick fix")
				fs.String("timeout", "1m", "max excution timeout")
				fs.String("out-format", "github-actions", "output format")
				fs.Int("concurrency", 1, "num of concurrent processes")
				return nil
			},
			Execute: func(ctx context.Context, r *readline.Readline) error {
				o.l.Info("Running golangci-lint ...")
				for _, dir := range o.dirs(ctx, r, "go.mod", 1) {
					o.l.Info("└  " + dir)
					if out, err := shell.New(ctx, o.l, "golangci-lint", "run").
						Args("--path-prefix", dir).
						Args(r.Flags()...).
						Args(r.PassThroughFlags()...).
						Args(r.AdditionalArgs()...).
						Dir(dir).
						Output(); err != nil {
						return errors.Wrap(err, string(out))
					}
				}
				return nil
			},
		})
	}
}

func CommandWithTSC() CommandOption {
	return func(o *Command) {
		o.commandTree.Nodes = append(o.commandTree.Nodes, &tree.Node{
			Name:        "tsc",
			Description: "run tsc",
			Args:        tree.Args{o.pathArg("tsconfig.json")},
			Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
				fs.Bool("fix", false, "run quick fix")
				return nil
			},
			Execute: func(ctx context.Context, r *readline.Readline) error {
				o.l.Info("Running tsc ...")
				for _, dir := range o.dirs(ctx, r, "tsconfig.json", 1) {
					o.l.Info("└  " + dir)

					if out, err := shell.New(ctx, o.l, "tsc", "--noEmit").
						Args(r.Flags()...).
						Args(r.PassThroughFlags()...).
						Args(r.AdditionalArgs()...).
						Dir(dir).
						Output(); err != nil {
						return errors.Wrap(err, string(out))
					}
				}
				return nil
			},
		})
	}
}

func CommandWithHelm() CommandOption {
	return func(o *Command) {
		o.commandTree.Nodes = append(o.commandTree.Nodes, &tree.Node{
			Name:        "helm",
			Description: "run helm lint",
			Args:        tree.Args{o.pathArg("Chart.yaml")},
			Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
				fs.Bool("fix", false, "run quick fix")
				return nil
			},
			Execute: func(ctx context.Context, r *readline.Readline) error {
				o.l.Info("Running helm ...")
				for _, dir := range o.dirs(ctx, r, "Chart.yaml", 1) {
					o.l.Info("└  " + dir)
					if out, err := shell.New(ctx, o.l, "helm", "lint", dir).
						Args(r.Flags()...).
						Args(r.PassThroughFlags()...).
						Args(r.AdditionalArgs()...).
						Output(); err != nil {
						return errors.Wrap(err, string(out))
					}
				}
				return nil
			},
		})
	}
}

func CommandWithESLint() CommandOption {
	return func(o *Command) {
		o.commandTree.Nodes = append(o.commandTree.Nodes, &tree.Node{
			Name:        "eslint",
			Description: "run eslint",
			Args:        tree.Args{o.pathArg("package.json")},
			Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
				fs.Bool("fix", false, "run quick fix")
				fs.Bool("cache", false, "use cache")
				return nil
			},
			Execute: func(ctx context.Context, r *readline.Readline) error {
				o.l.Info("Running eslint ...")
				for _, dir := range o.dirs(ctx, r, "package.json", 1) {
					o.l.Info("└  " + dir)
					if out, err := shell.New(ctx, o.l, "eslint", "--quiet", ".").
						Args(r.Flags()...).
						Args(r.PassThroughFlags()...).
						Args(r.AdditionalArgs()...).
						Dir(dir).
						Output(); err != nil {
						return errors.Wrap(err, string(out))
					}
				}
				return nil
			},
		})
	}
}

func CommandWithGherkin() CommandOption {
	return func(o *Command) {
		o.commandTree.Nodes = append(o.commandTree.Nodes, &tree.Node{
			Name:        "gherkin",
			Description: "run gherkin lint",
			Args:        tree.Args{o.pathArg("wdio.conf.ts")},

			Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
				fs.Bool("fix", false, "run quick fix")
				return nil
			},
			Execute: func(ctx context.Context, r *readline.Readline) error {
				o.l.Info("Running gherkin ...")
				for _, dir := range o.dirs(ctx, r, "wdio.conf.ts", 1) {
					o.l.Info("└  " + dir)

					if out, err := shell.New(ctx, o.l, "gherkin-lint", dir).
						Args(r.Flags()...).
						Args(r.PassThroughFlags()...).
						Args(r.AdditionalArgs()...).
						Output(); err != nil {
						return errors.Wrap(err, string(out))
					}
				}
				return nil
			},
		})
	}
}

func CommandWithTerraform() CommandOption {
	return func(o *Command) {
		o.commandTree.Nodes = append(o.commandTree.Nodes, &tree.Node{
			Name:        "terraform",
			Description: "run tflint lint",
			Args:        tree.Args{o.pathArg("main.tf")},
			Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
				fs.Bool("fix", false, "run quick fix")
				return nil
			},
			Execute: func(ctx context.Context, r *readline.Readline) error {
				o.l.Info("Running tflint ...")
				for _, dir := range o.dirs(ctx, r, "main.tf", 1) {
					o.l.Info("└  " + dir)

					if out, err := shell.New(ctx, o.l, "tflint").
						Dir(dir).
						Args(r.Flags()...).
						Args(r.PassThroughFlags()...).
						Args(r.AdditionalArgs()...).
						Output(); err != nil {
						return errors.Wrap(err, string(out))
					}
				}
				return nil
			},
		})
	}
}

func CommandWithTerrascan() CommandOption {
	return func(o *Command) {
		o.commandTree.Nodes = append(o.commandTree.Nodes, &tree.Node{
			Name:        "terrascan",
			Description: "run terrascan",
			Nodes: tree.Nodes{
				{
					Name:        "helm",
					Description: "run terrascan helm",
					Args:        tree.Args{o.pathArg("Chart.yaml")},
					Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
						fs.Bool("fix", false, "run quick fix")
						return nil
					},
					Execute: func(ctx context.Context, r *readline.Readline) error {
						o.l.Info("Running terrascan helm ...")
						for _, dir := range o.dirs(ctx, r, "Chart.yaml", 2) {
							o.l.Info("└  " + dir)
							if out, err := shell.New(ctx, o.l, "terrascan", "scan").
								Args("--iac-dir", dir).
								Args("--iac-type", "docker").
								Args(r.Flags()...).
								Args(r.PassThroughFlags()...).
								Args(r.AdditionalArgs()...).
								Output(); err != nil {
								return errors.Wrap(err, string(out))
							}
						}
						return nil
					},
				},
				{
					Name:        "terraform",
					Description: "run terrascan terraform",
					Args:        tree.Args{o.pathArg("main.tf")},
					Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
						fs.Bool("fix", false, "run quick fix")
						return nil
					},
					Execute: func(ctx context.Context, r *readline.Readline) error {
						o.l.Info("Running terrascan terraform ...")
						for _, dir := range o.dirs(ctx, r, "main.tf", 2) {
							o.l.Info("└  " + dir)
							if out, err := shell.New(ctx, o.l, "terrascan", "scan").
								Args("--iac-dir", dir).
								Args("--iac-type", "docker").
								Args(r.Flags()...).
								Args(r.PassThroughFlags()...).
								Args(r.AdditionalArgs()...).
								Output(); err != nil {
								return errors.Wrap(err, string(out))
							}
						}
						return nil
					},
				},
				{
					Name:        "docker",
					Description: "run terrascan docker",
					Args:        tree.Args{o.pathArg("Dockerfile")},
					Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
						fs.Bool("fix", false, "run quick fix")
						return nil
					},
					Execute: func(ctx context.Context, r *readline.Readline) error {
						o.l.Info("Running terrascan docker ...")
						for _, dir := range o.dirs(ctx, r, "Dockerfile", 2) {
							o.l.Info("└  " + dir)
							if out, err := shell.New(ctx, o.l, "terrascan", "scan").
								Args("--iac-dir", dir).
								Args("--iac-type", "docker").
								Args(r.Flags()...).
								Args(r.PassThroughFlags()...).
								Args(r.AdditionalArgs()...).
								Output(); err != nil {
								return errors.Wrap(err, string(out))
							}
						}
						return nil
					},
				},
			},
		})
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, c cache.Cache, opts ...CommandOption) *Command {
	inst := &Command{
		l:     l.Named("lint"),
		name:  "lint",
		cache: c.Get("lint"),
		commandTree: &tree.Root{
			Description: "lint your code",
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}
	inst.commandTree.Name = inst.name

	return inst
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.commandTree.Name
}

func (c *Command) Description() string {
	return c.commandTree.Description
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return `Lint your code.

Usage:
  lint [linter] <path>

Examples:
  lint go
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context, filename string) []string {
	return c.cache.Get("paths-"+strings.ToLower(filename), func() any {
		if value, err := files.Find(ctx, ".", filename, files.FindWithIgnore(`^\.`, "dist", "node_modules")); err != nil {
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

func (c *Command) pathArg(filename string) *tree.Arg {
	return &tree.Arg{
		Name:     "path",
		Optional: true,
		Suggest: func(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
			return suggests.List(c.paths(ctx, filename))
		},
	}
}

func (c *Command) dirs(ctx context.Context, r *readline.Readline, filename string, offset int) []string {
	var ret []string
	if r.Args().LenGt(offset) {
		for _, value := range r.Args()[offset:] {
			ret = append(ret, value)
		}
	} else {
		ret = c.paths(ctx, filename)
	}
	return ret
}
