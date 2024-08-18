package cloudflared

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"

	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
)

type (
	Command struct {
		l           log.Logger
		name        string
		cloudflared *Cloudflared
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

func NewCommand(l log.Logger, cloudflared *Cloudflared, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:           l.Named("cloudflared"),
		name:        "cloudflared",
		cloudflared: cloudflared,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := os.MkdirAll(inst.cloudflared.Config().Path, 0700); err != nil {
		return nil, err
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run cloudflared",
		Nodes: tree.Nodes{
			{
				Name:        "access",
				Description: "Forward access",
				Nodes: tree.Nodes{
					{
						Name:        "list",
						Description: "list forward access",
						Execute:     inst.accessList,
					},
					{
						Name:        "connect",
						Description: "open access by name ",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Name of the access",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.cloudflared.Config().AccessNames())
								},
							},
						},
						Execute: inst.accessConnect,
					},
					{
						Name:        "disconect",
						Description: "close access by name ",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Name of the access",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.cloudflared.Config().AccessNames())
								},
							},
						},
						Execute: inst.accessDisconnect,
					},
				},
				Execute: inst.execute,
			},
			{
				Name:        "tunnel",
				Description: "manage tunnels",
				Nodes: tree.Nodes{
					{
						Name:        "login",
						Description: "Generate a configuration file with your login details",
						Execute:     inst.execute,
					},
					{
						Name:        "create",
						Description: "Create a new tunnel with given name",
						Args: tree.Args{
							{
								Name:        "tunnel",
								Description: "UUID or name",
							},
						},
						Execute: inst.tunnelCreate,
					},
					{
						Name:        "delete",
						Description: "Delete existing tunnel by UUID or name",
						Args: tree.Args{
							{
								Name:        "tunnel",
								Description: "UUID or name",
							},
						},
						Execute: inst.execute,
					},
					{
						Name:        "route",
						Description: "Define which traffic routed from Cloudflare edge to this tunnel",
						Nodes: tree.Nodes{
							{
								Name:        "dns",
								Description: "HostnameRoute a hostname by creating a DNS CNAME record to a tunnel",
								Args: tree.Args{
									{
										Name:        "tunnel",
										Description: "UUID or name",
									},
									{
										Name:        "hostname",
										Description: "Hostname for the dns enty",
									},
								},
								Execute: inst.execute,
							},
						},
						Execute: inst.execute,
					},
					{
						Name:        "token",
						Description: "Create a new tunnel",
						Args: tree.Args{
							{
								Name:        "tunnel",
								Description: "UUID or name",
							},
						},
						Execute: inst.execute,
					},
					{
						Name:        "list",
						Description: "List existing tunnels",
						Execute:     inst.execute,
					},
					{
						Name:        "info",
						Description: "List details about the active connectors for a tunnel",
						Execute:     inst.execute,
					},
				},
				Execute: inst.execute,
			},
		},
		Execute: inst.execute,
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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	if _, err := exec.LookPath("cloudflared"); err != nil {
		c.l.Print()
		return errors.New("missing cloudflared executable")
	}
	return nil
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
	return shell.New(ctx, c.l, "cloudflared").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Env("HOME=" + c.cloudflared.Config().Path).
		Run()
}

func (c *Command) accessList(ctx context.Context, r *readline.Readline) error {
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

func (c *Command) accessConnect(ctx context.Context, r *readline.Readline) error {
	access := c.cloudflared.Config().GetAccesss(r.Args().At(2))
	return c.cloudflared.Connect(ctx, access)
}

func (c *Command) accessDisconnect(ctx context.Context, r *readline.Readline) error {
	access := c.cloudflared.Config().GetAccesss(r.Args().At(2))
	return c.cloudflared.Disonnect(ctx, access)
}

func (c *Command) tunnelCreate(ctx context.Context, r *readline.Readline) error {
	if err := shell.New(ctx, c.l, "cloudflared", "tunnel", "create", r.Args().At(2)).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Env("HOME=" + c.cloudflared.Config().Path).
		Run(); err != nil {
		return err
	}

	out, err := shell.New(ctx, c.l, "cloudflared", "tunnel", "token", r.Args().At(2)).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Env("HOME=" + c.cloudflared.Config().Path).
		Output()
	if err != nil {
		return err
	}

	var outDec []byte
	if _, err := base64.StdEncoding.Decode(outDec, out); err != nil {
		return err
	}

	c.l.Info(string(outDec))

	return nil
}
