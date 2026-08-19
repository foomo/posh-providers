package doctl

import (
	"context"
	_ "embed"

	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

//go:embed SKILL.md
var skill string

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
						Description: "Authenticate doctl by prompting for an API token",
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
						Description: "Write registry credentials into the host Docker config",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("never-expire", false, "Never expire the credentials")
							return nil
						},
						Execute: inst.exec,
					},
					{
						Name:        "logout",
						Description: "Remove registry credentials from the host Docker config",
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
						Description: "Save cluster credentials into a per-cluster kubeconfig under the kubectl config path",
						Args: tree.Args{
							{
								Name:        "cluster",
								Description: "Cluster to save credentials for, keyed as configured under clusters",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.doctl.cfg.ClusterNames())
								},
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Internal().String("profile", "", "Kubectl config profile to store the kubeconfig under")
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog cannot
// show that the registry verbs act on the host's Docker config rather than the
// project; that <cluster> is a local alias completing from config rather than
// from the API; that the written kubeconfig carries no credential and needs
// doctl on PATH to work at all; or that DIGITALOCEAN_ACCESS_TOKEN is never
// exported because the guard that would set it cannot be satisfied.
func (c *Command) Skill(ctx context.Context) string {
	return skill
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
