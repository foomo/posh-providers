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

type Command struct {
	l           log.Logger
	name        string
	cache       cache.Cache
	kubectl     *kubectl.Kubectl
	teleport    *Teleport
	commandTree tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, teleport *Teleport, kubectl *kubectl.Kubectl, opts ...Option) *Command {
	inst := &Command{
		l:        l.Named("teleport"),
		cache:    cache,
		kubectl:  kubectl,
		teleport: teleport,
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "teleport",
		Description: "Manage access points through teleport",
		Nodes: tree.Nodes{
			{
				Name:    "login",
				Execute: inst.login,
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
		},
	})

	/*
	 // Execute ...
	 func (c *Teleport) Execute(args, passArgs, pipeArgs []string) error {
	 	var env []string
	 	cmd, args := args[0], args[1:]

	 	env = append(env, "HOME="+path.Join(os.Getenv("PROJECT_ROOT"), "devops", "config"))

	 	switch cmd {
	 	case "login":
	 		if err := shell.New("tsh", "login").
	 			WithArgs(
	 				"--auth=github",
	 				fmt.Sprintf("--proxy=%s:443", c.config().Hostname),
	 			).
	 			WithPassArgs(passArgs...).
	 			WithPipeArgs(pipeArgs...).
	 			WithEnv(env).
	 			Run(); err != nil {
	 			return err
	 		}
	 	case "database":
	 		database, _ := args[0], args[1:]
	 		databaseUser := "developers"
	 		if value := os.Getenv("TELEPORT_DATABASE_USER"); value != "" {
	 			databaseUser = value
	 		}
	 		if err := shell.New("tsh", "db", "login", "--db-user", databaseUser, database).
	 			WithPassArgs(passArgs...).
	 			WithPipeArgs(pipeArgs...).
	 			WithEnv(env).
	 			Run(); err != nil {
	 			return err
	 		}
	 	case "kubeconfig":
	 		cluster, _ := args[0], args[1:]
	 		env = append(env, fmt.Sprintf("KUBECONFIG=%s", kubectl.Cluster(cluster).GetConfig()))
	 		// delete old config
	 		kubectl.Cluster(cluster).DeleteConfig()
	 		// generate & filter new config
	 		if err := shell.New("tsh", "kube", "login", cluster).
	 			WithPassArgs(passArgs...).
	 			WithPipeArgs(pipeArgs...).
	 			WithEnv(env).
	 			Run(); err != nil {
	 			return err
	 		} else if err = config.GenerateTeleportKubeconfig(kubectl.Cluster(cluster).GetConfig(), cluster); err != nil {
	 			return errors.Wrap(err, "failed to create teleport kubeconfig")
	 		}
	 	}

	 	cache.Clear("")

	 	return nil
	 }

	*/

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
		fmt.Sprintf("--proxy=%s", c.teleport.Config().Hostname),
		"--db-user", c.teleport.Config().Database.EnvUser(),
		databse,
	).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) kubeconfig(ctx context.Context, r *readline.Readline) error {
	cluster := c.kubectl.Cluster(r.Args().At(1))

	// delete old config
	if err := cluster.DeleteConfig(); err != nil {
		return err
	}

	// generate & filter new config
	return shell.New(ctx, c.l, "tsh", "kube", "login",
		fmt.Sprintf("--proxy=%s", c.teleport.Config().Hostname),
		cluster.Name(),
	).
		Env(cluster.Env()).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) login(ctx context.Context, r *readline.Readline) error {
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
