package pluto

import (
	"context"
	_ "embed"
	"sort"

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
		l:       l.Named("pluto"),
		kubectl: kubectl,
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "pluto",
		Description: "Find Kubernetes resources that have been deprecated",
		Nodes: tree.Nodes{
			{
				Name: "cluster",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					var ret []string
					for _, cluster := range inst.kubectl.Clusters() {
						ret = append(ret, cluster.Name())
					}

					sort.Strings(ret)

					return suggests.List(ret)
				},
				Description: "Cluster name",
				Nodes: tree.Nodes{
					{
						Name:        "detect",
						Description: "Scan all cluster resources for deprecated apiVersions",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("only-show-removed", false, "Only display the apiVersions that have been removed in the target version")
							fs.Default().String("output", "wide", "The output format to use. (normal|wide|custom|json|yaml|markdown|csv)")
							fs.Internal().String("profile", "", "Profile to use")

							if r.Args().HasIndex(0) {
								if err := fs.Internal().SetValues("profile", inst.kubectl.Cluster(r.Args().At(0)).Profiles(ctx)...); err != nil {
									return err
								}
							}

							return nil
						},
						Execute: inst.detect,
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
// show the three --ignore-* flags this provider appends unconditionally, which
// are exactly what makes `detect` exit 0 on a finding, nor that the cluster
// placeholder is resolved from kubeconfig files on disk.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) detect(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()
	ifs := r.FlagSets().Internal()
	cluster, args := c.kubectl.Cluster(r.Args().At(0)), r.Args()[2:]

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	return shell.New(ctx, c.l, "pluto", "detect-all-in-cluster").
		Args(args...).
		Args("--ignore-removals").
		Args("--ignore-deprecations").
		Args("--ignore-unavailable-replacements").
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Args("2>/dev/null").
		Env(cluster.Env(profile)).
		Run()
}
