package teleport

import (
	"context"
	"fmt"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/cache"
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
		name        string
		cache       cache.Cache
		kubectl     *kubectl.Kubectl
		teleport    *Teleport
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

func NewCommand(l log.Logger, cache cache.Cache, teleport *Teleport, kubectl *kubectl.Kubectl, opts ...CommandOption) *Command {
	inst := &Command{
		l:        l.Named("teleport"),
		name:     "teleport",
		cache:    cache,
		kubectl:  kubectl,
		teleport: teleport,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Manage access points through teleport",
		Execute:     inst.auth,
		Nodes: tree.Nodes{
			{
				Name:        "auth",
				Description: "Log in to a cluster and retrieve the session certificate",
				Execute:     inst.auth,
			},
			{
				Name:        "kubeconfig",
				Description: "Retrieve credentials to access remote cluster.",
				Args: tree.Args{
					{
						Name:        "cluster",
						Description: "Name of the cluster.",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.teleport.Clusters(ctx))
						},
					},
				},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("profile", "", "Profile to use.")
					return fs.Internal().SetValues("profile", "teleport")
				},
				Execute: inst.kubeconfig,
			},
			{
				Name:        "database",
				Description: "Retrieve credentials to access remote database.",
				Args: tree.Args{
					{
						Name:        "database",
						Description: "Name of the database.",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.teleport.Databases(ctx))
						},
					},
				},
				Execute: inst.database,
			},
			{
				Name:        "logout",
				Description: "Log out",
				Execute:     inst.logout,
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

func (c *Command) database(ctx context.Context, r *readline.Readline) error {
	databse := r.Args().At(1)

	return shell.New(ctx, c.l, "tsh", "db", "login",
		"--db-user", c.teleport.Config().Database.EnvUser(),
		databse,
	).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) kubeconfig(ctx context.Context, r *readline.Readline) error {
	ifs := r.FlagSets().Internal()
	cluster := c.kubectl.Cluster(r.Args().At(1))

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	// delete old config
	if err := cluster.DeleteConfig(profile); err != nil {
		return err
	}

	// generate & filter new config
	return shell.New(ctx, c.l, "tsh", "kube", "login",
		c.teleport.cfg.Kubernetes.Name(cluster.Name()),
	).
		Env(cluster.Env(profile)).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) auth(ctx context.Context, r *readline.Readline) error {
	if err := shell.New(ctx, c.l, "tsh", "login",
		fmt.Sprintf("--proxy=%s", c.teleport.Config().Hostname),
		"--auth=github",
	).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run(); err != nil {
		return err
	}
	return nil
}

func (c *Command) logout(ctx context.Context, r *readline.Readline) error {
	if err := shell.New(ctx, c.l, "tsh", "logout").
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run(); err != nil {
		return err
	}
	return nil
}
