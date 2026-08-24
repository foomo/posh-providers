package k9s

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/foomo/posh-providers/foomo/squadron"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
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

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

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
		l:       l.Named("k9s"),
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
			Description: "Cluster to open the dashboard against; sets KUBECONFIG, not validated",
			Suggest:     inst.completeClusters,
		},
	}
	if inst.squadron != nil {
		args = append(args, &tree.Arg{
			Name:        "fleet",
			Description: "Fleet within the cluster; omit to cover every namespace",
			Optional:    true,
			Suggest:     inst.completeFleets,
		},
			&tree.Arg{
				Name:        "squadron",
				Description: "Squadron within the fleet; combined with the fleet into the namespace",
				Optional:    true,
				Suggest:     inst.completeSquadrons,
			})
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "k9s",
		Description: "Open the k9s dashboard",
		Args:        args,
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	switch {
	case r.Args().LenIs(0):
		return errors.New("missing [cluster] argument")
	case !c.kubectl.Cluster(r.Args().At(0)).ConfigExistsForFlags(r.Flags()):
		return errors.Errorf("invalid [cluster] argument: %s", r.Args().At(0))
	}

	return nil
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

// Skill implements the optional command.Skiller interface. The catalog shows a
// single node with up to three arguments and cannot show that it opens a
// blocking, read-write terminal UI; that `[fleet]` and
// `[squadron]` exist only when the project wired CommandWithSquadron; or how
// the two combine into a namespace - including the case where `all` yields a
// literal `<fleet>-all` rather than every namespace.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the interactive-dashboard phrasings a human asks for, since the first
// thing the file has to say is that an agent must not run this at all.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when a human wants an interactive Kubernetes dashboard or TUI for a " +
			"cluster in this project - browsing, editing, scaling or deleting live resources " +
			"from the terminal, or \"open the cluster UI\" - and when explaining why such a " +
			"dashboard cannot be driven unattended. Also for scoping one to a squadron " +
			"fleet's namespace.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	var args []string

	ifs := r.FlagSets().Internal()
	cluster, fleet, squad := r.Args().At(0), r.Args().AtDefault(1, ""), r.Args().AtDefault(2, "")

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	env := []string{c.kubectl.Cluster(cluster).Env(profile)}

	if value := c.squadronNamespaceFn(cluster, fleet, squad); value == "all" {
		args = append(args, "--all-namespaces")
	} else if value != "" {
		args = append(args, "--namespace="+value)
	}

	return shell.New(ctx, c.l, "k9s", "--logoless", "--splashless").
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
