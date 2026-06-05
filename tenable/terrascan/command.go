package terrascan

import (
	"context"
	"path"

	"github.com/foomo/go/options"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/exec"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

const (
	modeHelm      = "helm"
	modeTerraform = "terraform"
	modeDocker    = "docker"
)

type Command struct {
	l             log.Logger
	name          string
	cache         cache.Namespace
	execTerrascan exec.CommandProvider
	commandTree   tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithName(v string) options.Option[*Command] {
	return func(o *Command) {
		o.name = v
	}
}

func CommandWithExecTerrascan(v exec.CommandProvider) options.Option[*Command] {
	return func(o *Command) {
		o.execTerrascan = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, c cache.Cache, opts ...options.Option[*Command]) *Command {
	inst := &Command{
		l:     l.Named("terrascan"),
		name:  "terrascan",
		cache: c.Get("terrascan"),
		execTerrascan: func(ctx context.Context, args ...string) *exec.Command {
			return exec.NewCommand(ctx, "terrascan", args...)
		},
	}

	options.Apply(inst, opts...)

	pathArg := func(filename string) *tree.Arg {
		return &tree.Arg{
			Name:     "path",
			Optional: true,
			Repeat:   true,
			Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
				return suggests.List(inst.paths(ctx, filename))
			},
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run terrascan",
		Nodes: tree.Nodes{
			{
				Name:        "helm",
				Description: "Scan helm charts",
				Args:        tree.Args{pathArg("Chart.yaml")},
				Execute:     inst.executeHelm,
			},
			{
				Name:        "terraform",
				Description: "Scan terraform modules",
				Args:        tree.Args{pathArg("main.tf")},
				Execute:     inst.executeTerraform,
			},
			{
				Name:        "docker",
				Description: "Scan dockerfiles",
				Args:        tree.Args{pathArg("Dockerfile")},
				Execute:     inst.executeDocker,
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

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

func (c *Command) Lint(ctx context.Context, _ bool) error {
	for _, mode := range []string{modeHelm, modeTerraform, modeDocker} {
		if err := c.scan(ctx, mode, nil); err != nil {
			return err
		}
	}

	return nil
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) executeHelm(ctx context.Context, r *readline.Readline) error {
	return c.scan(ctx, modeHelm, r.Args().From(1))
}

func (c *Command) executeTerraform(ctx context.Context, r *readline.Readline) error {
	return c.scan(ctx, modeTerraform, r.Args().From(1))
}

func (c *Command) executeDocker(ctx context.Context, r *readline.Readline) error {
	return c.scan(ctx, modeDocker, r.Args().From(1))
}

func (c *Command) scan(ctx context.Context, mode string, paths []string) error {
	filename, err := filenameFor(mode)
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		paths = c.paths(ctx, filename)
	}

	c.l.Info("Running terrascan " + mode + " ...")

	for _, dir := range paths {
		c.l.Info("└ " + dir)

		if err := c.execTerrascan(ctx, "scan", "--iac-dir", dir, "--iac-type", mode).Run(); err != nil {
			return err
		}
	}

	return nil
}

//nolint:forcetypeassert
func (c *Command) paths(ctx context.Context, filename string) []string {
	return c.cache.Get("paths-"+filename, func() any {
		matches, err := files.Find(ctx, ".", filename, files.FindWithIgnore(`^(node_modules|\.\w*)$`))
		if err != nil {
			c.l.Debug("failed to walk files", err.Error())
			return []string{}
		}

		out := make([]string, 0, len(matches))
		for _, m := range matches {
			out = append(out, path.Dir(m))
		}

		return out
	}).([]string)
}

func filenameFor(mode string) (string, error) {
	switch mode {
	case modeHelm:
		return "Chart.yaml", nil
	case modeTerraform:
		return "main.tf", nil
	case modeDocker:
		return "Dockerfile", nil
	default:
		return "", errors.Errorf("unknown terrascan mode: %s", mode)
	}
}
