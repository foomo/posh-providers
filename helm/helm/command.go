package helm

import (
	"context"
	_ "embed"
	"strings"

	"github.com/foomo/go/options"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/exec"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
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
	kubectl     *kubectl.Kubectl
	execHelm    exec.CommandProvider
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

func CommandWithExecHelm(v exec.CommandProvider) options.Option[*Command] {
	return func(o *Command) {
		o.execHelm = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, kubectl *kubectl.Kubectl, opts ...options.Option[*Command]) *Command {
	inst := &Command{
		l:       l.Named("helm"),
		name:    "helm",
		kubectl: kubectl,
		execHelm: func(ctx context.Context, args ...string) *exec.Command {
			return exec.NewCommand(ctx, "helm", args...)
		},
	}

	options.Apply(inst, opts...)

	allFlags := func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
		fs.Default().Bool("help", false, "help for helm")
		fs.Default().Bool("debug", false, "enable verbose output")
		fs.Default().String("namespace", "", "namespace scope for this request")
		fs.Default().Bool("all-namespaces", false, "all namespace scope for this request")
		fs.Default().Bool("create-namespace", false, "create the release namespace if not present")
		fs.Default().Bool("dependency-update", false, "update dependencies")
		fs.Default().Bool("dry-run", false, "simulate the operation without applying any change")
		fs.Default().Bool("atomic", false, "delete installation on failure")
		fs.Default().Bool("wait", false, "wait until all resources a ready")
		fs.Internal().String("profile", "", "Subdirectory of kubectl's config path to read the cluster kubeconfig from")

		if r.Args().HasIndex(0) {
			if err := fs.Internal().SetValues("profile", inst.kubectl.Cluster(r.Args().At(0)).Profiles(ctx)...); err != nil {
				return err
			}
		}

		return nil
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run helm against a cluster, using its kubeconfig",
		Nodes: tree.Nodes{
			{
				Name: "cluster",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.kubectl.Clusters())
				},
				Description: "Cluster name, selecting the kubeconfig every subcommand runs with",
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
								Name:        "value",
								Description: "Which part of the release to download",
								Repeat:      false,
								Optional:    false,
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
	case !c.kubectl.Cluster(r.Args().At(0)).ConfigExistsForFlags(r.Flags()):
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The rendered tree
// lists every subcommand without saying which mutate the cluster, that the
// cluster argument is the only thing scoping them, or that this tree is a
// hand-maintained mirror rather than a passthrough.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the release-level verbs, because the destructive ones are the reason
// this file needs to be loaded before the command is run.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when working with Helm releases on a Kubernetes cluster in this " +
			"project - installing, upgrading, rolling back or uninstalling a release, listing " +
			"releases, checking release status or history, rendering or diffing chart " +
			"templates, or linting and packaging a chart. Also for chart repositories and " +
			"\"which chart version is deployed\".",
	}
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

	return c.execHelm(ctx).
		Args(args...).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Env(cluster.Env(profile)).
		Run()
}
