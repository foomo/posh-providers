package beam

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
)

type (
	Command struct {
		l           log.Logger
		beam        *Beam
		name        string
		kubectl     *kubectl.Kubectl
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

func NewCommand(l log.Logger, beam *Beam, kubectl *kubectl.Kubectl, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:       l.Named("beam"),
		name:    "beam",
		beam:    beam,
		kubectl: kubectl,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run beam",
		Nodes: tree.Nodes{
			{
				Name: "tunnel",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.beam.cfg.GetTunnelNames())
				},
				Description: "Tunnel",
				Nodes: tree.Nodes{
					{
						Name:        "kubeconfig",
						Description: "Download kubeconfig",
						Args: tree.Args{
							{
								Name:        "cluster",
								Description: "Cluster name",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.beam.Config().GetTunnel(r.Args().At(0)).GetClusterNames())
								},
							},
						},
						Execute: inst.kubeconfig,
					},
				},
			},
		},
	})

	return inst, nil
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

func (c *Command) kubeconfig(ctx context.Context, r *readline.Readline) error {
	tunnel := c.beam.Config().GetTunnel(r.Args().At(0))
	cluster := tunnel.GetCluster(r.Args().At(2))
	kubeconfig, err := c.beam.op.GetDocument(ctx, cluster.Credentials)
	if err != nil {
		return err
	}

	filename := path.Join(c.kubectl.Config().ConfigPath, r.Args().At(2)+".yaml")
	c.l.Info("Retrieving kubeconfig", "tunnel", r.Args().At(0), "cluster", r.Args().At(2), "filename", filename)

	kubeconfig = strings.ReplaceAll(kubeconfig, "$PORT", fmt.Sprintf("%d", cluster.Port))

	if err := os.WriteFile(filename, []byte(kubeconfig), 0600); err != nil {
		return err
	}
	return nil
}
