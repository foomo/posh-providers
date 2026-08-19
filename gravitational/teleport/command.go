package teleport

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
)

//go:embed SKILL.md
var skill string

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
		Description: "Manage access points through teleport; logs in when called without a subcommand",
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
				Name:        "app",
				Description: "Retrieve credentials to access remote app.",
				Args: tree.Args{
					{
						Name:        "name",
						Description: "Name of the app",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.teleport.Apps(ctx))
						},
					},
				},
				Execute: inst.app,
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog lists
// five verbs and cannot show that all of them need an interactive browser SSO
// an agent cannot complete; that the bare root logs in rather than printing
// help; that `kubeconfig` deletes the existing config before the login that
// would replace it; that nothing completes until authenticated, and then only
// what the configured labels match; or that cluster aliases are reverse-mapped
// by scanning a map, so a duplicated or shadowing alias picks a cluster at
// random.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) app(ctx context.Context, r *readline.Readline) error {
	app := r.Args().At(1)
	appArgs := c.teleport.Config().Apps[app]

	return shell.New(ctx, c.l, "tsh", "apps", "login").
		Args(appArgs...).
		Args(app).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

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
