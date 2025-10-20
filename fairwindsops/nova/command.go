package nova

import (
	"context"
	"sort"

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
		l:       l.Named("nova"),
		kubectl: kubectl,
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "nova",
		Description: "Find outdated or deprecated Helm charts running in your cluster",
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
						Name:        "find",
						Description: "",
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("containers", false, "Show old container image versions instead of helm chart versions")
							fs.Default().Bool("helm", false, "Show old helm chart versions")
							fs.Default().Bool("include-all", false, "Show all charts even if no latest version is found")
							fs.Default().Bool("show-errored-containers", false, "Show errors encountered when scanning")
							fs.Default().Bool("show-non-semver", false, "Show all containers even if they don't follow semver")
							fs.Default().Bool("show-old", false, "Only show charts that are not on the latest version")
							fs.Default().Bool("wide", false, "Output chart name and namespace")
							fs.Default().String("config", "", "Config file to use")
							fs.Default().String("format", "table", "An output format (table, json)")
							fs.Default().String("namespace", "", "Namespace to look in")
							fs.Internal().String("profile", "", "Profile to use")

							if r.Args().HasIndex(0) {
								if err := fs.Internal().SetValues("profile", inst.kubectl.Cluster(r.Args().At(0)).Profiles(ctx)...); err != nil {
									return err
								}
							}

							return nil
						},
						Execute: inst.execute,
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

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()
	ifs := r.FlagSets().Internal()
	cluster, args := c.kubectl.Cluster(r.Args().At(0)), r.Args()[1:]

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	return shell.New(ctx, c.l, "nova").
		Args(args...).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Env(cluster.Env(profile)).
		Run()
}
