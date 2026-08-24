package kubeprompt

import (
	"context"
	_ "embed"
	"strings"

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

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

type (
	Command struct {
		l           log.Logger
		kubectl     *kubectl.Kubectl
		commandTree tree.Root
	}
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, kubectl *kubectl.Kubectl, opts ...CommandOption) *Command {
	inst := &Command{
		l:       l.Named("kubeprompt"),
		kubectl: kubectl,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "kubeprompt",
		Description: "Open the interactive kube-prompt shell against a cluster",
		Args: tree.Args{
			{
				Name:        "cluster",
				Description: "Cluster name, selecting the kubeconfig kube-prompt runs with",
				Suggest:     inst.completeClusters,
			},
		},
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			if r.Args().HasIndex(0) {
				fs.Internal().String("profile", "", "Subdirectory of kubectl's config path to read the cluster kubeconfig from")

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

// Skill implements the optional command.Skiller interface. The rendered tree
// cannot show that this opens a blocking interactive shell unusable unattended,
// that the whole kubectl surface is reachable inside it against the selected
// cluster, or that the binary is `kube-prompt` while the command is
// `kubeprompt`.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the interactive-prompt request and the binary's real name, since the
// only thing an agent needs from this file is that it must not run it.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when a human wants an interactive kubectl prompt with completion " +
			"against a cluster in this project - an exploratory kubectl session rather than " +
			"one-shot commands - and when explaining why such a session cannot be driven " +
			"unattended. Also for the kube-prompt binary and \"kube-prompt: command not found\".",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	ifs := r.FlagSets().Internal()

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	cluster := r.Args().At(0)
	env := []string{c.kubectl.Cluster(cluster).Env(profile)}

	return shell.New(ctx, c.l, "kube-prompt").
		Args(r.AdditionalArgs()...).
		Env(env...).
		Run()
}

func (c *Command) completeClusters(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	return suggests.List(c.kubectl.Clusters())
}
