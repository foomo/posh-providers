package ku

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/foomo/posh-providers/foomo/squadron"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
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
			Name:        "cluster",
			Description: "Cluster to connect to; selects the kubeconfig",
			Suggest:     inst.completeClusters,
		},
	}
	if inst.squadron != nil {
		args = append(args, &tree.Arg{
			Name:        "fleet",
			Description: "Fleet to scope the namespace to; every namespace if omitted",
			Optional:    true,
			Suggest:     inst.completeFleets,
		},
			&tree.Arg{
				Name:        "squadron",
				Description: "Squadron to narrow the namespace to within the fleet",
				Optional:    true,
				Suggest:     inst.completeSquadrons,
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog renders
// `<cluster> [fleet] [squadron]` and cannot show that the last two exist only
// when the project passed CommandWithSquadron, that they are folded into a
// single `--namespace` by a function whose four branches drop the fleet name in
// two of them, or that omitting `[fleet]` is cluster-wide rather than a default
// namespace. It also cannot show that this is a blocking full-screen TUI, or
// that `--edit` turns it read-write against a cluster argument nothing
// validates.
func (c *Command) Skill(ctx context.Context) string {
	return skill
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
