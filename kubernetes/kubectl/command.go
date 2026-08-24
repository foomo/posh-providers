package kubectl

import (
	"context"
	_ "embed"
	"strings"

	"github.com/foomo/go/options"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/exec"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
)

//go:embed SKILL.md
var skill string

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

type Command struct {
	l           log.Logger
	kubectl     *Kubectl
	execKubectl exec.CommandProvider
	commandTree tree.Root
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithExecKubectl(v exec.CommandProvider) options.Option[*Command] {
	return func(c *Command) {
		c.execKubectl = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, kubectl *Kubectl, opts ...options.Option[*Command]) *Command {
	inst := &Command{
		l:       l.Named("kubectl"),
		kubectl: kubectl,
		execKubectl: func(ctx context.Context, args ...string) *exec.Command {
			return exec.NewCommand(ctx, "kubectl", args...)
		},
	}

	options.Apply(inst, opts...)

	inst.commandTree = tree.New(&tree.Node{
		Name:        "kubectl",
		Description: "Run kubectl against a cluster, selecting its kubeconfig by name",
		Args: tree.Args{
			{
				Name:        "cluster",
				Description: "Cluster name, matching a <cluster>.yaml kubeconfig in the configured path",
				Suggest:     inst.completeClusters,
			},
		},
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			if r.Args().HasIndex(0) {
				fs.Internal().String("profile", "", "Subdirectory of the config path to read the cluster's kubeconfig from")

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
// shows one argument, not that the whole kubectl CLI follows it, that this
// provider only selects among kubeconfigs other providers wrote, that
// `--profile` redirects which file is read rather than reaching kubectl, or
// that listing clusters chmods them to 0600.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the raw kubectl verbs and the kubeconfig symptoms an agent would phrase
// a request with, since this is the command every other cluster verb falls back
// to and the runtime matches on that phrasing alone.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when inspecting or changing Kubernetes resources in this project - " +
			"getting pods, describing or logging a workload, applying or deleting a manifest, " +
			"exec into a container - and when a cluster must be named rather than taken from " +
			"the current context. Also for \"which clusters are configured\", KUBECONFIG " +
			"questions, kubeconfig profiles, and \"no configuration has been provided\" errors.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	profile, err := r.FlagSets().Internal().GetString("profile")
	if err != nil {
		return err
	}

	cluster := c.kubectl.Cluster(r.Args().At(0))

	return c.execKubectl(ctx, r.Args().From(1)...).
		Args(r.AdditionalFlags()...).
		Args(r.AdditionalArgs()...).
		Env(cluster.Env(profile)).
		Run()
}

func (c *Command) completeClusters(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	return suggests.List(c.kubectl.Clusters())
}
