package ssh

import (
	"context"
	_ "embed"

	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

//go:embed SKILL.md
var skill string

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
		Description: "Manage ssh port forwards and socks5 tunnels",
		Nodes: []*tree.Node{
			{
				Name:        "pfw",
				Description: "Manage ssh local port forwards (ssh -L)",
				Nodes: []*tree.Node{
					{
						Name:        "start",
						Description: "Start configured port forwards as background processes; all of them if no name is given",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Port forward name from the portForwards config; all of them if omitted",
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
						Description: "Stop running port forwards; all of them if no name is given",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Port forward name from the portForwards config; all of them if omitted",
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
				Description: "Manage ssh socks5 tunnels (ssh -D)",
				Nodes: []*tree.Node{
					{
						Name:        "start",
						Description: "Start configured socks5 tunnels as background processes; all of them if no name is given",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Socks5 tunnel name from the socks5Tunnels config; all of them if omitted",
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
						Description: "Stop running socks5 tunnels; all of them if no name is given",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Socks5 tunnel name from the socks5Tunnels config; all of them if omitted",
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The rendered tree
// cannot show that these verbs manage background processes outliving the shell,
// that an omitted name makes start/stop act on every configured entry, that
// `pfw start` blocks on an interactive auth prompt where `socks5 start` does
// not, or that `port: 0` fails to auto-assign for a port forward.
func (c *Command) Skill(ctx context.Context) string {
	return skill
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
