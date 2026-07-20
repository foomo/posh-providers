package ku

import (
	"context"
	"fmt"

	"github.com/foomo/posh-providers/foomo/squadron"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
)

type (
	Command struct {
		l                   log.Logger
		kubectl             *kubectl.Kubectl
		squadron            squadron.Squadron
		commandTree         tree.Root
		squadronNamespaceFn NamespaceFn
	}
	NamespaceFn   func(cluster, fleet, squadron string) string
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithSquadron(v squadron.Squadron) CommandOption {
	return func(o *Command) {
		o.squadron = v
	}
}

func CommandWithSquadronNamespaceFn(v NamespaceFn) CommandOption {
	return func(o *Command) {
		o.squadronNamespaceFn = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, kubectl *kubectl.Kubectl, opts ...CommandOption) *Command {
	inst := &Command{
		l:       l.Named("ku"),
		kubectl: kubectl,
		squadronNamespaceFn: func(cluster, fleet, squadron string) string {
			switch {
			case fleet == "":
				return ""
			case squadron == "":
				return fleet
			case fleet == "default":
				return squadron
			default:
				return fmt.Sprintf("%s-%s", fleet, squadron)
			}
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	args := tree.Args{
		{
			Name:    "cluster",
			Suggest: inst.completeClusters,
		},
	}
	if inst.squadron != nil {
		args = append(args, &tree.Arg{
			Name:     "fleet",
			Optional: true,
			Suggest:  inst.completeFleets,
		},
			&tree.Arg{
				Name:     "squadron",
				Optional: true,
				Suggest:  inst.completeSquadrons,
			})
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "ku",
		Description: "Open the ku dashboard",
		Args:        args,
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			fs.Default().Bool("edit", false, "Start in edit mode (default is read-only).")
			fs.Default().Bool("dev", false, "Developer view, app resources only.")

			if r.Args().HasIndex(0) {
				fs.Internal().String("profile", "", "Profile to use.")

				if err := fs.Internal().SetValues("profile", inst.kubectl.Cluster(r.Args().At(0)).Profiles(ctx)...); err != nil {
					return err
				}
			}

			return nil
		},
		Execute: inst.execute,
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

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	var args []string

	ifs := r.FlagSets().Internal()
	dfs := r.FlagSets().Default()
	cluster, fleet, squad := r.Args().At(0), r.Args().AtDefault(1, ""), r.Args().AtDefault(2, "")

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	edit, err := dfs.GetBool("edit")
	if err != nil {
		return err
	}

	dev, err := dfs.GetBool("dev")
	if err != nil {
		return err
	}

	env := []string{c.kubectl.Cluster(cluster).Env(profile)}

	if value := c.squadronNamespaceFn(cluster, fleet, squad); value != "" && value != "all" {
		args = append(args, "--namespace="+value)
	}

	if edit {
		args = append(args, "--edit")
	}

	if dev {
		args = append(args, "--dev")
	}

	return shell.New(ctx, c.l, "ku").
		Env(env...).
		Args(args...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) completeClusters(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
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
