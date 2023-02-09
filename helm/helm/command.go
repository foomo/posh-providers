package helm

import (
	"context"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

type (
	Command struct {
		l           log.Logger
		name        string
		kubectl     *kubectl.Kubectl
		commandTree *tree.Root
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

func NewCommand(l log.Logger, kubectl *kubectl.Kubectl) *Command {
	inst := &Command{
		l:       l.Named("helm"),
		name:    "helm",
		kubectl: kubectl,
	}
	allFlags := func(fs *readline.FlagSet) {
		fs.Bool("help", false, "help for helm")
		fs.Bool("debug", false, "enable verbose output")
		fs.String("namespace", "", "namespace scope for this request")
		fs.Bool("all-namespaces", false, "all namespace scope for this request")
		fs.Bool("create-namespace", false, "create the release namespace if not present")
		fs.Bool("dependency-update", false, "update dependencies")
		fs.Bool("dry-run", false, "assume aws profile")
		fs.Bool("atomic", false, "delete installation on failure")
		fs.Bool("wait", false, "wait until all resources a ready")
	}

	inst.commandTree = &tree.Root{
		Name:        inst.name,
		Description: "run helm commands",
		Nodes: tree.Nodes{
			{
				Name: "cluster",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.kubectl.Clusters())
				},
				Description: "cluster to run against",
				Nodes: tree.Nodes{
					{
						Name:        "create",
						Description: "create a new chart with the given name",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "dependency",
						Description: "manage a chart's dependencies",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "diff",
						Description: "preview helm upgrade changes as a diff",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "env",
						Description: "helm client environment information",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "get",
						Description: "download extended information of a named release",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							fs.String("revision", "", "get the named release with revision")
							return nil
						},
						Args: tree.Args{
							{
								Name:     "value",
								Repeat:   false,
								Optional: false,
								Suggest: func(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
									return []goprompt.Suggest{
										{Text: "all", Description: "download all information for a named release"},
										{Text: "hooks", Description: "download all hooks for a named release"},
										{Text: "manifest", Description: "download the manifest for a named release"},
										{Text: "notes", Description: "download the notes for a named release"},
										{Text: "values", Description: "download the values file for a named release"},
									}
								},
							},
						},
						Execute: inst.execute,
					},
					{
						Name:        "help",
						Description: "Help about any command",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "history",
						Description: "fetch release history",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "install",
						Description: "install a chart",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "lint",
						Description: "examine a chart for possible issues",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "list",
						Description: "list releases",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "package",
						Description: "package a chart directory into a chart archive",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "plugin",
						Description: "install, list, or uninstall Helm plugins",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "pull",
						Description: "download a chart from a repository and (optionally) unpack it in local directory",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "repo",
						Description: "add, list, remove, update, and index chart repositories",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "rollback",
						Description: "roll back a release to a previous revision",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "search",
						Description: "search for a keyword in charts",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "show",
						Description: "show information of a chart",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "status",
						Description: "display the status of the named release",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							fs.Bool("show-desc", false, "show description")
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "template",
						Description: "locally render templates",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "test",
						Description: "run tests for a release",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "uninstall",
						Description: "uninstall a release",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "upgrade",
						Description: "upgrade a release",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "verify",
						Description: "verify that a chart at the given path has been signed and is valid",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "version",
						Description: "print the client version information",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
							allFlags(fs)
							return nil
						},
						Execute: inst.execute,
					},
				},
			},
		},
	}

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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	switch {
	case r.Args().LenIs(0):
		return errors.New("missing [CLUSTER] argument")
	case !c.kubectl.Cluster(r.Args().At(0)).ConfigExists():
		return errors.New("invalid [CLUSTER] argument")
	case r.Args().LenIs(1):
		return errors.New("missing [CMD] argument")
	}

	return nil
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return `Generate helm files.

Usage:
  helm <path> <options>

Available options:
  debug
  skiphelm

Examples:
  helm ./path/helm.yml
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	cluster, args := c.kubectl.Cluster(r.Args().At(0)), r.Args()[1:]

	return shell.New(ctx, c.l, "helm").
		Args(args...).
		Args(r.Flags()...).
		Args(r.PassThroughFlags()...).
		Args(r.AdditionalArgs()...).
		Env(cluster.Env()).
		Run()
}
