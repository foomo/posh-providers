package beam

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/foomo/posh-providers/cloudflare/cloudflared"
	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pterm/pterm"
)

type (
	Command struct {
		l           log.Logger
		beam        *Beam
		name        string
		kubectl     *kubectl.Kubectl
		cloudflared *cloudflared.Cloudflared
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

func NewCommand(l log.Logger, beam *Beam, kubectl *kubectl.Kubectl, cloudflared *cloudflared.Cloudflared, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:           l.Named("beam"),
		name:        "beam",
		beam:        beam,
		kubectl:     kubectl,
		cloudflared: cloudflared,
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
				Name:        "status",
				Description: "Show connection status",
				Execute:     inst.status,
			},
			{
				Name:        "cluster",
				Description: "Manage cluster connection",
				Nodes: tree.Nodes{
					{
						Name:        "name",
						Description: "Cluster name",
						Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.beam.cfg.ClusterNames())
						},
						Nodes: tree.Nodes{
							{
								Name:        "connect",
								Description: "Connect to cluster",
								Execute:     inst.clusterConnect,
							},
							{
								Name:        "kubeconfig",
								Description: "Download kubeconfig",
								Execute:     inst.clusterKubeconfig,
							},
							{
								Name:        "disconnect",
								Description: "Disconnect to cluster",
								Execute:     inst.clusterDisconnect,
							},
						},
						Execute: nil,
					},
				},
			},
			{
				Name:        "database",
				Description: "Manage cluster connection",
				Nodes: tree.Nodes{
					{
						Name:        "name",
						Description: "Cluster name",
						Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.beam.cfg.DatabaseNames())
						},
						Nodes: tree.Nodes{
							{
								Name:        "connect",
								Description: "Connect to database",
								Execute:     inst.databaseConnect,
							},
							{
								Name:        "disconnect",
								Description: "Disconnect to database",
								Execute:     inst.databaseDisconnect,
							},
						},
						Execute: nil,
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

func (c *Command) status(ctx context.Context, r *readline.Readline) error {
	list, err := c.cloudflared.List()
	if err != nil {
		return err
	}

	data := pterm.TableData{
		{"pid", "cmdline"},
	}

	for _, p := range list {
		data = append(data, []string{p.PID, p.Cmdline})
	}

	return pterm.DefaultTable.WithHasHeader(true).WithData(data).Render()
}

func (c *Command) clusterKubeconfig(ctx context.Context, r *readline.Readline) error {
	clusterName := r.Args().At(1)
	clusterConfig := c.beam.Config().GetCluster(clusterName)
	kubectlCluster := c.kubectl.Cluster(clusterName)

	c.l.Info("Retrieving kubeconfig", "cluster", clusterName, "filename", kubectlCluster.Config(""))

	kubeconfig, err := c.beam.op.GetDocument(ctx, clusterConfig.Kubeconfig)
	if err != nil {
		return err
	}
	kubeconfig = strings.ReplaceAll(kubeconfig, "$PORT", fmt.Sprintf("%d", clusterConfig.Port))

	return os.WriteFile(kubectlCluster.Config(""), []byte(kubeconfig), 0600)
}

func (c *Command) clusterConnect(ctx context.Context, r *readline.Readline) error {
	clusterName := r.Args().At(1)
	clusterConfig := c.beam.Config().GetCluster(clusterName)

	return c.cloudflared.Connect(ctx, cloudflared.Access{
		Type:     "tcp",
		Hostname: clusterConfig.Hostname,
		Port:     clusterConfig.Port,
	})
}

func (c *Command) clusterDisconnect(ctx context.Context, r *readline.Readline) error {
	clusterName := r.Args().At(1)
	clusterConfig := c.beam.Config().GetCluster(clusterName)

	return c.cloudflared.Disonnect(ctx, cloudflared.Access{
		Type:     "tcp",
		Hostname: clusterConfig.Hostname,
		Port:     clusterConfig.Port,
	})
}

func (c *Command) databaseConnect(ctx context.Context, r *readline.Readline) error {
	databaseName := r.Args().At(1)
	databaseConfig := c.beam.Config().GetDatabase(databaseName)

	return c.cloudflared.Connect(ctx, cloudflared.Access{
		Type:     "tcp",
		Hostname: databaseConfig.Hostname,
		Port:     databaseConfig.Port,
	})
}

func (c *Command) databaseDisconnect(ctx context.Context, r *readline.Readline) error {
	databaseName := r.Args().At(1)
	databaseConfig := c.beam.Config().GetDatabase(databaseName)

	return c.cloudflared.Disonnect(ctx, cloudflared.Access{
		Type:     "tcp",
		Hostname: databaseConfig.Hostname,
		Port:     databaseConfig.Port,
	})
}
