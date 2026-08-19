package beam

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/foomo/posh-providers/cloudflare/cloudflared"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pterm/pterm"
)

//go:embed SKILL.md
var skill string

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
		Description: "Manage cloudflared tunnels to remote clusters and databases",
		Nodes: tree.Nodes{
			{
				Name:        "status",
				Description: "List every running cloudflared process, not just this project's",
				Execute:     inst.status,
			},
			{
				Name:        "cluster",
				Description: "Manage cluster tunnels",
				Nodes: tree.Nodes{
					{
						Name:        "connect",
						Description: "Open a cloudflared tunnel to the cluster",
						Args: tree.Args{
							{
								Name:        "cluster",
								Description: "Cluster name from the clusters config",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.beam.cfg.ClusterNames())
								},
							},
						},
						Execute: inst.clusterConnect,
					},
					{
						Name:        "kubeconfig",
						Description: "Fetch the kubeconfig from 1Password and overwrite kubectl's file for this cluster",
						Args: tree.Args{
							{
								Name:        "cluster",
								Description: "Cluster name from the clusters config",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.beam.cfg.ClusterNames())
								},
							},
						},
						Execute: inst.clusterKubeconfig,
					},
					{
						Name:        "disconnect",
						Description: "Close cluster tunnels by killing the cloudflared process; all clusters if no name is given",
						Args: tree.Args{
							{
								Name:        "cluster",
								Description: "Cluster name from the clusters config; all of them if omitted",
								Optional:    true,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.beam.cfg.ClusterNames())
								},
							},
						},
						Execute: inst.clusterDisconnect,
					},
				},
			},
			{
				Name:        "database",
				Description: "Manage database tunnels",
				Nodes: tree.Nodes{
					{
						Name:        "connect",
						Description: "Open a cloudflared tunnel to the database",
						Args: tree.Args{
							{
								Name:        "database",
								Description: "Database name from the databases config",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.beam.cfg.DatabaseNames())
								},
							},
						},
						Execute: inst.databaseConnect,
					},
					{
						Name:        "disconnect",
						Description: "Close database tunnels by killing the cloudflared process; all databases if no name is given",
						Args: tree.Args{
							{
								Name:        "database",
								Description: "Database name from the databases config; all of them if omitted",
								Optional:    true,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.beam.cfg.DatabaseNames())
								},
							},
						},
						Execute: inst.databaseDisconnect,
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The rendered tree
// cannot show that `disconnect` with no name closes every tunnel and matches on
// hostname alone, that `kubeconfig` overwrites kubectl's file from 1Password,
// that an unknown name silently resolves to an empty config rather than an
// error, or that the gokazi tasks this provider registers are never used.
func (c *Command) Skill(ctx context.Context) string {
	return skill
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
	clusterName := r.Args().At(2)
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
	name := r.Args().At(2)
	clusterConfig := c.beam.Config().GetCluster(name)

	c.l.Info("Connecting to cluster: " + name)

	return c.cloudflared.Connect(ctx, cloudflared.Access{
		Type:     "tcp",
		Hostname: clusterConfig.Hostname,
		Port:     clusterConfig.Port,
	})
}

func (c *Command) clusterDisconnect(ctx context.Context, r *readline.Readline) error {
	var names []string
	if r.Args().LenGt(2) {
		names = r.Args().From(2)
	} else {
		names = c.beam.cfg.ClusterNames()
	}

	for _, name := range names {
		c.l.Info("Disconnecting from cluster: " + name)

		clusterConfig := c.beam.Config().GetCluster(name)
		if err := c.cloudflared.Disconnect(ctx, cloudflared.Access{
			Type:     "tcp",
			Hostname: clusterConfig.Hostname,
			Port:     clusterConfig.Port,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) databaseConnect(ctx context.Context, r *readline.Readline) error {
	name := r.Args().At(2)
	databaseConfig := c.beam.Config().GetDatabase(name)

	c.l.Info("Connecting to database: " + name)

	return c.cloudflared.Connect(ctx, cloudflared.Access{
		Type:     "tcp",
		Hostname: databaseConfig.Hostname,
		Port:     databaseConfig.Port,
	})
}

func (c *Command) databaseDisconnect(ctx context.Context, r *readline.Readline) error {
	var names []string
	if r.Args().LenGt(2) {
		names = r.Args().From(2)
	} else {
		names = c.beam.cfg.DatabaseNames()
	}

	for _, name := range names {
		c.l.Info("Disconnecting from database: " + name)

		databaseConfig := c.beam.Config().GetDatabase(name)
		if err := c.cloudflared.Disconnect(ctx, cloudflared.Access{
			Type:     "tcp",
			Hostname: databaseConfig.Hostname,
			Port:     databaseConfig.Port,
		}); err != nil {
			return err
		}
	}

	return nil
}
