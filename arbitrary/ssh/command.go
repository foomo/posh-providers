package ssh

import (
	"context"

	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

type (
	Command struct {
		l           log.Logger
		ssh         *SSH
		name        string
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

func NewCommand(l log.Logger, ssh *SSH, opts ...CommandOption) *Command {
	inst := &Command{
		l:    l.Named("ssh"),
		name: "ssh",
		ssh:  ssh,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Manage ssh",
		Nodes: []*tree.Node{
			{
				Name:        "pfw",
				Description: "Manage port forwards",
				Nodes: []*tree.Node{
					{
						Name:        "start",
						Description: "Start a port forwaring",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Config name",
								Repeat:      true,
								Optional:    true,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.ssh.Config().PortForwardNames())
								},
							},
						},
						Execute: inst.startPortForward,
					},
					{
						Name:        "stop",
						Description: "Stop a port forwaring",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Config name",
								Repeat:      true,
								Optional:    true,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.ssh.Config().PortForwardNames())
								},
							},
						},
						Execute: inst.stopPortForward,
					},
				},
			},
			{
				Name:        "socks5",
				Description: "Manage socks5 tunnels",
				Nodes: []*tree.Node{
					{
						Name:        "start",
						Description: "Start a socks5 tunnel",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Config name",
								Repeat:      true,
								Optional:    true,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.ssh.Config().Socks5TunnelNames())
								},
							},
						},
						Execute: inst.startSocks5Tunnel,
					},
					{
						Name:        "stop",
						Description: "Stop a socks5 tunnel",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Config name",
								Repeat:      true,
								Optional:    true,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.ssh.Config().Socks5TunnelNames())
								},
							},
						},
						Execute: inst.stopSocks5Tunnel,
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

func (c *Command) startPortForward(ctx context.Context, r *readline.Readline) error {
	names := c.ssh.Config().PortForwardNames()
	if r.Args().LenGt(2) {
		names = r.Args().From(2)
	}

	for _, value := range names {
		v, ok := c.ssh.Config().PortForward(value)
		if !ok {
			return errors.Errorf("%s not found", value)
		}

		c.l.Infof("Starting port forward %d:%s:%d", v.Port, v.Host, v.HostPort)

		if err := c.ssh.StartPortForward(ctx, value); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) stopPortForward(ctx context.Context, r *readline.Readline) error {
	names := c.ssh.Config().PortForwardNames()
	if r.Args().LenGt(2) {
		names = r.Args().From(2)
	}

	for _, name := range names {
		v, ok := c.ssh.Config().PortForward(name)
		if !ok {
			return errors.Errorf("%s not found", name)
		}

		c.l.Infof("Stopping port forward %d:%s:%d", v.Port, v.Host, v.HostPort)

		if err := c.ssh.StopPortForward(ctx, name); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) startSocks5Tunnel(ctx context.Context, r *readline.Readline) error {
	names := c.ssh.Config().Socks5TunnelNames()
	if r.Args().LenGt(2) {
		names = r.Args().From(2)
	}

	for _, value := range names {
		v, ok := c.ssh.Config().Socks5Tunnel(value)
		if !ok {
			return errors.Errorf("%s not found", value)
		}

		c.l.Infof("Starting SOCK proxy %d:%s:%d", v.Port, v.Host, v.HostPort)

		if err := c.ssh.StartSocks5Tunnel(ctx, value); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) stopSocks5Tunnel(ctx context.Context, r *readline.Readline) error {
	names := c.ssh.Config().Socks5TunnelNames()
	if r.Args().LenGt(2) {
		names = r.Args().From(2)
	}

	for _, name := range names {
		v, ok := c.ssh.Config().Socks5Tunnel(name)
		if !ok {
			return errors.Errorf("%s not found", name)
		}

		c.l.Infof("Stopping port forward %d:%s:%d", v.Port, v.Host, v.HostPort)

		if err := c.ssh.StopSocks5Tunnel(ctx, name); err != nil {
			return err
		}
	}

	return nil
}
