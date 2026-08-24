package stern

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/foomo/posh-providers/foomo/squadron"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
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
		name        string
		cfg         Config
		configKey   string
		kubectl     *kubectl.Kubectl
		squadron    squadron.Squadron
		commandTree tree.Root
		namespaceFn NamespaceFn
	}
	NamespaceFn   func(cluster, fleet, squadron string) string
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

func CommandWithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

func CommandWithNamespaceFn(v NamespaceFn) CommandOption {
	return func(o *Command) {
		o.namespaceFn = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, kubectl *kubectl.Kubectl, squadron squadron.Squadron, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("stern"),
		name:      "stern",
		configKey: "stern",
		kubectl:   kubectl,
		squadron:  squadron,
		namespaceFn: func(cluster, fleet, squadron string) string {
			if fleet == "default" {
				return squadron
			} else {
				return fmt.Sprintf("%s-%s", fleet, squadron)
			}
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "stern",
		Description: "Tail Kubernetes pod logs with stern",
		Nodes: []*tree.Node{
			{
				Name:        "cluster",
				Description: "Cluster whose kubeconfig is used to tail",
				Values:      inst.completeClusters,
				Nodes: []*tree.Node{
					{
						Name:        "query",
						Description: "Tail using a named query from this provider's config",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Query names, walked as a path into the nested config; each level's arguments are appended",
								Repeat:      true,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.cfg.QueryNames(r.Args().From(2)...))
								},
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("only-log-lines", false, "Print only log lines")
							fs.Default().Int("tail", -1, "The number of lines from the end of the logs to show")
							fs.Default().Bool("all-namespaces", false, "If present, tail across all namespaces")
							fs.Default().String("namespace", "", "Kubernetes namespace to use")
							fs.Default().String("container", "", "Container name when multiple containers in pod (default \".*\")")
							fs.Default().String("exclude", "", "Regex of log lines to exclude")
							fs.Default().String("exclude-container", "", "Exclude a Container name")
							fs.Default().String("include", "", "Regex of log lines to include")
							fs.Default().String("output", "default", "Specify predefined template")
							fs.Default().String("selector", "", "Selector (label query) to filter on. If present, default to \".*\" for the pod-query")
							fs.Default().String("since", "default", "Return logs newer than a relative duration like 5s, 2m, or 3")
							fs.Default().String("template", "default", "Template to use for log lines")
							fs.Internal().String("profile", "", "Kubectl config profile holding this cluster's kubeconfig")

							if err := fs.Default().SetValues("output", "raw", "json", "extjson", "ppextjson"); err != nil {
								return err
							}

							if r.Args().HasIndex(0) {
								if err := fs.Internal().SetValues("profile", inst.kubectl.Cluster(r.Args().At(0)).Profiles(ctx)...); err != nil {
									return err
								}

								if err := fs.Default().SetValues("namespace", inst.kubectl.Cluster(r.Args().At(0)).Namespaces(ctx, "")...); err != nil {
									return err
								}
							}

							return nil
						},
						Execute: inst.tailQuery,
					},
					{
						Name:        "raw",
						Description: "Tail using a pod regex given directly",
						Args: tree.Args{
							{
								Name:        "query",
								Description: "Pod name regex passed straight to stern",
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("only-log-lines", false, "Print only log lines")
							fs.Default().Int("tail", -1, "The number of lines from the end of the logs to show")
							fs.Default().Bool("all-namespaces", false, "If present, tail across all namespaces")
							fs.Default().String("namespace", "", "Kubernetes namespace to use")
							fs.Default().String("container", "", "Container name when multiple containers in pod (default \".*\")")
							fs.Default().String("exclude", "", "Regex of log lines to exclude")
							fs.Default().String("exclude-container", "", "Exclude a Container name")
							fs.Default().String("include", "", "Regex of log lines to include")
							fs.Default().String("output", "default", "Specify predefined template")
							fs.Default().String("selector", "", "Selector (label query) to filter on. If present, default to \".*\" for the pod-query")
							fs.Default().String("since", "default", "Return logs newer than a relative duration like 5s, 2m, or 3")
							fs.Default().String("template", "default", "Template to use for log lines")
							fs.Internal().String("profile", "", "Kubectl config profile holding this cluster's kubeconfig")

							if err := fs.Default().SetValues("output", "raw", "json", "extjson", "ppextjson"); err != nil {
								return err
							}

							if r.Args().HasIndex(0) {
								if err := fs.Internal().SetValues("profile", inst.kubectl.Cluster(r.Args().At(0)).Profiles(ctx)...); err != nil {
									return err
								}

								if err := fs.Default().SetValues("namespace", inst.kubectl.Cluster(r.Args().At(0)).Namespaces(ctx, "")...); err != nil {
									return err
								}
							}

							return nil
						},
						Execute: inst.tailRaw,
					},
					{
						Name:        "squadron",
						Description: "Tail a squadron unit's pods, deriving the namespace from fleet and squadron",
						Args: tree.Args{
							{
								Name:        "fleet",
								Description: "Squadron fleet, used to derive the namespace",
								Suggest:     inst.completeFleets,
							},
							{
								Name:        "squadron",
								Description: "Squadron name, used for both the namespace and the pod prefix",
								Suggest:     inst.completeSquadrons,
							},
							{
								Name:        "unit",
								Description: "Squadron unit; matched as pods named <squadron>-<unit>",
								Suggest:     inst.completeSquadronUnits,
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("only-log-lines", false, "Print only log lines")
							fs.Default().Int("tail", -1, "The number of lines from the end of the logs to show")
							fs.Default().String("container", "", "Container name when multiple containers in pod (default \".*\")")
							fs.Default().String("exclude", "", "Regex of log lines to exclude")
							fs.Default().String("exclude-container", "", "Exclude a Container name")
							fs.Default().String("include", "", "Regex of log lines to include")
							fs.Default().String("output", "default", "Specify predefined template")
							fs.Default().String("selector", "", "Selector (label query) to filter on. If present, default to \".*\" for the pod-query")
							fs.Default().String("since", "default", "Return logs newer than a relative duration like 5s, 2m, or 3")
							fs.Default().String("template", "default", "Template to use for log lines")
							fs.Internal().String("profile", "", "Kubectl config profile holding this cluster's kubeconfig")

							if err := fs.Default().SetValues("output", "raw", "json", "extjson", "ppextjson"); err != nil {
								return err
							}

							if r.Args().HasIndex(0) {
								if err := fs.Internal().SetValues("profile", inst.kubectl.Cluster(r.Args().At(0)).Profiles(ctx)...); err != nil {
									return err
								}
							}

							return nil
						},
						Execute: inst.tailSquadron,
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

// Skill implements the optional command.Skiller interface. The catalog renders
// [name]... as one repeated argument and cannot show that it is a path into a
// recursive query tree whose levels concatenate, that an unmatched name after a
// valid one is silently dropped, that stern streams until interrupted, or that
// `squadron` computes its own namespace.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names log-reading phrasings rather than the tool, since "why is this service
// erroring" is how a request for a multi-pod tail actually arrives.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when reading Kubernetes pod logs in this project - tailing or " +
			"following the logs of a deployment, a squadron unit or every pod matching a " +
			"pattern, across containers and namespaces at once, or grepping logs for an error " +
			"while debugging a failing or crash-looping workload. Also for this project's " +
			"named, predefined log queries.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) tail(ctx context.Context, r *readline.Readline, args ...string) error {
	cluster := r.Args().At(0)
	fs := r.FlagSets().Default()
	ifs := r.FlagSets().Internal()

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	cmd := shell.New(ctx, c.l, "stern").
		Env(c.kubectl.Cluster(cluster).Env(profile)).
		Args(args...).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...)

	return cmd.Run()
}

func (c *Command) tailQuery(ctx context.Context, r *readline.Readline) error {
	queries := c.cfg.FindQueries(r.Args().From(2)...)
	if queries == nil {
		return errors.New("query not found")
	}

	var args []string
	for _, query := range queries {
		args = append(args, query.Query...)
	}

	return c.tail(ctx, r, args...)
}

func (c *Command) tailRaw(ctx context.Context, r *readline.Readline) error {
	return c.tail(ctx, r, r.Args().At(2))
}

func (c *Command) tailSquadron(ctx context.Context, r *readline.Readline) error {
	cluster, fleet, squad, unit := r.Args().At(0), r.Args().At(2), r.Args().At(3), r.Args().At(4)
	return c.tail(ctx, r, squad+"-"+unit, "--namespace", c.namespaceFn(cluster, fleet, squad))
}

func (c *Command) completeClusters(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return suggests.List(c.kubectl.Clusters())
}

func (c *Command) completeFleets(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	if cluster, ok := c.squadron.Cluster(r.Args().At(0)); ok {
		return suggests.List(cluster.Fleets)
	}

	return nil
}

func (c *Command) completeSquadrons(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	if value, err := c.squadron.List(); err != nil {
		c.l.Debug(err.Error())
		return nil
	} else {
		return suggests.List(value)
	}
}

func (c *Command) completeSquadronUnits(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	cluster, fleet, squad := r.Args().At(0), r.Args().At(2), r.Args().At(3)
	if value, err := c.squadron.ListUnits(ctx, squad, cluster, fleet, true); err != nil {
		c.l.Debug(err.Error())
		return nil
	} else {
		return suggests.List(value)
	}
}
