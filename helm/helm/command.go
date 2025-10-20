package helm

import (
	"context"

	"github.com/foomo/posh-providers/kubernetes/kubectl"
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

func NewCommand(l log.Logger, kubectl *kubectl.Kubectl) *Command {
	inst := &Command{
		l:       l.Named("helm"),
		name:    "helm",
		kubectl: kubectl,
	}

	allFlags := func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
		fs.Default().Bool("help", false, "help for helm")
		fs.Default().Bool("debug", false, "enable verbose output")
		fs.Default().String("namespace", "", "namespace scope for this request")
		fs.Default().Bool("all-namespaces", false, "all namespace scope for this request")
		fs.Default().Bool("create-namespace", false, "create the release namespace if not present")
		fs.Default().Bool("dependency-update", false, "update dependencies")
		fs.Default().Bool("dry-run", false, "assume aws profile")
		fs.Default().Bool("atomic", false, "delete installation on failure")
		fs.Default().Bool("wait", false, "wait until all resources a ready")
		fs.Internal().String("profile", "", "Profile to use.")

		if r.Args().HasIndex(0) {
			if err := fs.Internal().SetValues("profile", inst.kubectl.Cluster(r.Args().At(0)).Profiles(ctx)...); err != nil {
				return err
			}
		}

		return nil
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run helm commands",
		Nodes: tree.Nodes{
			{
				Name: "cluster",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.kubectl.Clusters())
				},
				Description: "Cluster to run against",
				Nodes: tree.Nodes{
					{
						Name:        "create",
						Description: "Create a new chart with the given name",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "dependency",
						Description: "Manage a chart's dependencies",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "diff",
						Description: "Preview helm upgrade changes as a diff",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "env",
						Description: "Helm client environment information",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "get",
						Description: "Download extended information of a named release",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().String("revision", "", "get the named release with revision")
							return allFlags(ctx, r, fs)
						},
						Args: tree.Args{
							{
								Name:     "value",
								Repeat:   false,
								Optional: false,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return []goprompt.Suggest{
										{Text: "all", Description: "Download all information for a named release"},
										{Text: "hooks", Description: "Download all hooks for a named release"},
										{Text: "manifest", Description: "Download the manifest for a named release"},
										{Text: "notes", Description: "Download the notes for a named release"},
										{Text: "values", Description: "Download the values file for a named release"},
									}
								},
							},
						},
						Execute: inst.execute,
					},
					{
						Name:        "help",
						Description: "Help about any command",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "history",
						Description: "Fetch release history",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "install",
						Description: "Install a chart",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "lint",
						Description: "Examine a chart for possible issues",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "list",
						Description: "List releases",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "package",
						Description: "Package a chart directory into a chart archive",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "plugin",
						Description: "Install, list, or uninstall Helm plugins",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "pull",
						Description: "Download a chart from a repository and (optionally) unpack it in local directory",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "repo",
						Description: "Add, list, remove, update, and index chart repositories",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "rollback",
						Description: "Roll back a release to a previous revision",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "search",
						Description: "Search for a keyword in charts",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "show",
						Description: "Show information of a chart",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "status",
						Description: "Display the status of the named release",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("show-desc", false, "show description")
							return allFlags(ctx, r, fs)
						},
						Execute: inst.execute,
					},
					{
						Name:        "template",
						Description: "Locally render templates",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "test",
						Description: "Run tests for a release",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "uninstall",
						Description: "Uninstall a release",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "upgrade",
						Description: "Upgrade a release",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "verify",
						Description: "Verify that a chart at the given path has been signed and is valid",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "version",
						Description: "Print the client version information",
						Flags:       allFlags,
						Execute:     inst.execute,
					},
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
	switch {
	case r.Args().LenIs(0):
		return errors.New("missing [CLUSTER] argument")
	case !c.kubectl.Cluster(r.Args().At(0)).ConfigExists(""):
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
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()
	ifs := r.FlagSets().Internal()
	cluster, args := c.kubectl.Cluster(r.Args().At(0)), r.Args()[1:]

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	return shell.New(ctx, c.l, "helm").
		Args(args...).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Env(cluster.Env(profile)).
		Run()
}
