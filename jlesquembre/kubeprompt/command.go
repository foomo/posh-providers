package kubeprompt

import (
	"context"

	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/pkg/proxy"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
		kubectl     *kubectl.Kubectl
		proxyCfg    proxy.Config
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

	_ = viper.UnmarshalKey("proxies", &inst.proxyCfg)

	inst.commandTree = tree.New(&tree.Node{
		Name:        "kubeprompt",
		Description: "Open the kubectl prompt",
		Args: tree.Args{
			{
				Name:    "cluster",
				Suggest: inst.completeClusters,
			},
		},
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
	ifs := r.FlagSets().Internal()

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	cluster := r.Args().At(0)
	env := []string{c.kubectl.Cluster(cluster).Env(profile)}

	if proxyName := c.kubectl.Config().ClusterProxy(cluster); proxyName != "" {
		proxyEnv, stop, err := c.proxyCfg.Start(ctx, c.l, proxyName)
		if err != nil {
			return err
		}
		defer stop()

		env = append(env, proxyEnv...)
	}

	return shell.New(ctx, c.l, "kube-prompt").
		Args(r.AdditionalArgs()...).
		Env(env...).
		Run()
}

func (c *Command) completeClusters(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	return suggests.List(c.kubectl.Clusters())
}
