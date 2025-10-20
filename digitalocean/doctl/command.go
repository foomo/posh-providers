package doctl

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
		doctl         *Doctl
		cache         cache.Cache
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

func NewCommand(l log.Logger, cache cache.Cache, doctl *Doctl, kubectl *kubectl.Kubectl, opts ...CommandOption) *Command {
	inst := &Command{
		l:       l.Named("doctl"),
		name:    "doctl",
		cache:   cache,
		doctl:   doctl,
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
		Description: "Manage digital ocean resources",
		Nodes: tree.Nodes{
			{
				Name:        "auth",
				Description: "Manage authentication",
				Nodes: tree.Nodes{
					{
						Name:        "init",
						Description: "Initialize doctl to use a specific account",
						Execute:     inst.exec,
					},
				},
				Execute: inst.exec,
			},
			{
				Name:        "registry",
				Description: "Manage container registries",
				Nodes: tree.Nodes{
					{
						Name:        "login",
						Description: "Log in Docker to a container registry",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("never-expire", false, "Never expire the credentials")
							return nil
						},
						Execute: inst.exec,
					},
					{
						Name:        "logout",
						Description: "Log out Docker from a container registry",
						Execute:     inst.exec,
					},
				},
				Execute: inst.exec,
			},
			{
				Name:        "kubernetes",
				Description: "Manage Kubernetes clusters",
				Nodes: tree.Nodes{
					{
						Name:        "kubeconfig",
						Description: "Retrieve credentials to access remote cluster.",
						Args: tree.Args{
							{
								Name:        "cluster",
								Description: "Name of the cluster.",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.doctl.cfg.ClusterNames())
								},
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Internal().String("profile", "", "Store credentials in given profile.")
							return fs.Internal().SetValues("profile", "digitalocean")
						},
						Execute: inst.kubeconfig,
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
	var args []string

	ifs := r.FlagSets().Internal()
	clusterName := r.Args().At(2)

	cluster, err := c.doctl.cfg.Cluster(clusterName)
	if err != nil {
		return errors.Errorf("failed to retrieve cluster for: %q", clusterName)
	}

	kubectlCluster := c.kubectl.Cluster(c.clusterNameFn(clusterName, cluster))
	if kubectlCluster == nil {
		return errors.Errorf("failed to retrieve kubectl cluster for: %q", cluster.Name)
	}

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	return shell.New(ctx, c.l, "doctl", "kubernetes", "cluster", "kubeconfig", "save", cluster.Name).
		Args(args...).
		Args(r.AdditionalArgs()...).
		Env(kubectlCluster.Env(profile)).
		Run()
}

func (c *Command) exec(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "doctl").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}
