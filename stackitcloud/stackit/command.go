package stackit

import (
	"context"

	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/cache"
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
		l             log.Logger
		name          string
		cache         cache.Namespace
		stackit       *Stackit
		kubectl       *kubectl.Kubectl
		commandTree   tree.Root
		clusterNameFn ClusterNameFn
	}
	ClusterNameFn func(name string, cluster Cluster) string
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

func CommandWithClusterNameFn(v ClusterNameFn) CommandOption {
	return func(o *Command) {
		o.clusterNameFn = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, stackit *Stackit, kubectl *kubectl.Kubectl, opts ...CommandOption) *Command {
	inst := &Command{
		l:       l.Named("stackit"),
		name:    "stackit",
		cache:   cache.Get("stackit"),
		stackit: stackit,
		kubectl: kubectl,
		clusterNameFn: func(name string, cluster Cluster) string {
			return name
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Manage stackit cloud resources",
		Nodes: tree.Nodes{
			{
				Name:        "auth",
				Description: "Login to the stackit cloud provider",
				Execute:     inst.auth,
			},
			{
				Name:        "project",
				Description: "Run a command against the given project",
				Nodes: tree.Nodes{
					{
						Name: "project",
						Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.stackit.Config().ProjectNames())
						},
						Description: "Project to run against",
						Nodes: tree.Nodes{
							{
								Name:        "kubeconfig",
								Description: "Retrieve credentials to access remote cluster.",
								Args: tree.Args{
									{
										Name:        "cluster",
										Description: "Name of the cluster.",
										Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
											project, err := inst.stackit.Config().Project(r.Args().At(1))
											if err != nil {
												return []goprompt.Suggest{}
											}

											return suggests.List(project.ClusterNames())
										},
									},
								},
								Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
									fs.Internal().String("profile", "", "Store credentials in given profile.")
									return fs.Internal().SetValues("profile", "stackit")
								},
								Execute: inst.kubeconfig,
							},
						},
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

func (c *Command) kubeconfig(ctx context.Context, r *readline.Readline) error {
	ifs := r.FlagSets().Internal()

	project, err := c.stackit.Config().Project(r.Args().At(1))
	if err != nil {
		return err
	}

	clusterName := r.Args().At(3)

	cluster, err := project.Cluster(clusterName)
	if err != nil {
		return err
	}

	kubectlCluster := c.kubectl.Cluster(c.clusterNameFn(clusterName, cluster))
	if kubectlCluster == nil {
		return errors.Errorf("failed to retrieve kubectl cluster for: %q", cluster.Name)
	}

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	return shell.New(ctx, c.l, "stackit", "ske", "kubeconfig", "create", cluster.Name).
		Args("--filepath", kubectlCluster.Config(profile)).
		Args("--project-id", project.ID).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) auth(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "stackit", "auth", "login").
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}
