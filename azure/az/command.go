package az

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
		l             log.Logger
		name          string
		az            *AZ
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

func NewCommand(l log.Logger, az *AZ, kubectl *kubectl.Kubectl, opts ...CommandOption) *Command {
	inst := &Command{
		l:       l.Named("az"),
		name:    "az",
		az:      az,
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
		Description: "Manage azure resources",
		Nodes: tree.Nodes{
			{
				Name:        "login",
				Description: "Log in to Azure",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Default().String("tenant", "", "The Microsoft Entra tenant")
					return nil
				},
				Execute: inst.exec,
			},
			{
				Name:        "logout",
				Description: "Log out to remove access to Azure subscriptions",
				Execute:     inst.exec,
			},
			{
				Name:        "configure",
				Description: "Manage Azure CLI configuration",
				Execute:     inst.exec,
			},
			{
				Name:        "artifactory",
				Description: "Login into the artifactory",
				Args: tree.Args{
					{
						Name:        "subscription",
						Description: "Name of the subscription",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.az.cfg.SubscriptionNames())
						},
					},
					{
						Name:        "artifactory",
						Description: "Name of the artifactory",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							ret, err := inst.az.cfg.Subscription(r.Args().At(1))
							if err != nil {
								return nil
							}
							return suggests.List(ret.ArtifactoryNames())
						},
					},
				},
				Execute: inst.artifactory,
			},
			{
				Name:        "kubeconfig",
				Description: "Retrieve credentials to access remote cluster",
				Args: tree.Args{
					{
						Name:        "subscription",
						Description: "Name of the subscription",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.az.cfg.SubscriptionNames())
						},
					},
					{
						Name:        "cluster",
						Description: "Name of the cluster",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							ret, err := inst.az.cfg.Subscription(r.Args().At(1))
							if err != nil {
								return nil
							}
							return suggests.List(ret.ClusterNames())
						},
					},
				},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("profile", "", "Store credentials in given profile")
					return fs.Internal().SetValues("profile", "azure")
				},
				Execute: inst.kubeconfig,
			},
		},
		Execute: inst.exec,
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

func (c *Command) artifactory(ctx context.Context, r *readline.Readline) error {
	sub, err := c.az.cfg.Subscription(r.Args().At(1))
	if err != nil {
		return errors.Errorf("failed to retrieve subscription for: %q", r.Args().At(1))
	}

	acr, err := sub.Artifactory(r.Args().At(2))
	if err != nil {
		return errors.Errorf("failed to retrieve artifactoy for: %q", r.Args().At(2))
	}

	if err := shell.New(ctx, c.l, "az", "acr", "login").
		Args("--name", acr.Name).
		Args("--resource-group", acr.ResourceGroup).
		Args("--subscription", sub.Name).
		Run(); err != nil {
		return err
	}

	return nil
}
func (c *Command) kubeconfig(ctx context.Context, r *readline.Readline) error {
	ifs := r.FlagSets().Internal()

	sub, err := c.az.cfg.Subscription(r.Args().At(1))
	if err != nil {
		return errors.Errorf("failed to retrieve subscription for: %q", r.Args().At(1))
	}

	k8s, err := sub.Cluster(r.Args().At(2))
	if err != nil {
		return errors.Errorf("failed to retrieve cluster for: %q", r.Args().At(2))
	}

	kubectlCluster := c.kubectl.Cluster(c.clusterNameFn(r.Args().At(2), k8s))
	if kubectlCluster == nil {
		return errors.Errorf("failed to retrieve kubectl cluster for: %q", k8s.Name)
	}

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	if err := shell.New(ctx, c.l, "az", "aks", "get-credentials").
		Args("--name", k8s.Name).
		Args("--resource-group", k8s.ResourceGroup).
		Args("--subscription", sub.Name).
		Args("--overwrite-existing").
		Env(kubectlCluster.Env(profile)).
		Run(); err != nil {
		return err
	}

	return shell.New(ctx, c.l, "kubelogin", "convert-kubeconfig", "-l", "azurecli").
		Env(kubectlCluster.Env(profile)).
		Run()
}

func (c *Command) exec(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "az").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}
