package stern

import (
	"context"
	"fmt"

	"github.com/foomo/posh-providers/foomo/squadron"
	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
)

type (
	Command struct {
		l           log.Logger
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

func CommandWithNamespaceFn(v NamespaceFn) CommandOption {
	return func(o *Command) {
		o.namespaceFn = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, kubectl *kubectl.Kubectl, squadron squadron.Squadron, opts ...CommandOption) *Command {
	inst := &Command{
		l:        l.Named("stern"),
		kubectl:  kubectl,
		squadron: squadron,
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
	inst.commandTree = tree.New(&tree.Node{
		Name:        "stern",
		Description: "Tail your logs with stern",
		Nodes: []*tree.Node{
			{
				Name:        "cluster",
				Description: "Cluster name",
				Values:      inst.completeClusters,
				Nodes: []*tree.Node{
					{
						Name:        "query",
						Description: "Tail by query",
						Args: tree.Args{
							{
								Name: "query",
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("only-log-lines", false, "Print only log lines")
							fs.Default().Int("tail", -1, "The number of lines from the end of the logs to show")
							fs.Default().String("all-namespaces", "", "If present, tail across all namespaces")
							fs.Default().String("namespace", "", "Kubernetes namespace to use")
							fs.Default().String("container", "", "Container name when multiple containers in pod (default \".*\")")
							fs.Default().String("exclude", "", "Regex of log lines to exclude")
							fs.Default().String("exclude-container", "", "Exclude a Container name")
							fs.Default().String("include", "", "Regex of log lines to include")
							fs.Default().String("output", "default", "Specify predefined template")
							fs.Default().String("selector", "", "Selector (label query) to filter on. If present, default to \".*\" for the pod-query.")
							fs.Default().String("since", "default", "Return logs newer than a relative duration like 5s, 2m, or 3")
							fs.Default().String("template", "default", "Template to use for log lines")
							fs.Internal().String("profile", "", "Profile to use.")
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
						Name:        "squadron",
						Description: "Tail by squadron unit",
						Args: tree.Args{
							{
								Name:    "fleet",
								Suggest: inst.completeFleets,
							},
							{
								Name:    "squadron",
								Suggest: inst.completeSquadrons,
							},
							{
								Name:    "unit",
								Suggest: inst.completeSquadronUnits,
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
							fs.Default().String("selector", "", "Selector (label query) to filter on. If present, default to \".*\" for the pod-query.")
							fs.Default().String("since", "default", "Return logs newer than a relative duration like 5s, 2m, or 3")
							fs.Default().String("template", "default", "Template to use for log lines")
							fs.Internal().String("profile", "", "Profile to use")
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

	return shell.New(ctx, c.l, "stern").
		Env(c.kubectl.Cluster(cluster).Env(profile)).
		Args(args...).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) tailQuery(ctx context.Context, r *readline.Readline) error {
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
