package cloudflared

import (
	"context"
	_ "embed"
	"encoding/base64"
	"os"
	"os/exec"
	"strings"

	"github.com/foomo/posh/pkg/agent"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
)

//go:embed SKILL.md
var skill string

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

type (
	// Command is the posh `cloudflared` command, managing Access connections
	// and account-level tunnels.
	Command struct {
		l           log.Logger
		name        string
		cloudflared *Cloudflared
		commandTree tree.Root
	}
	// CommandOption configures a Command in NewCommand.
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// CommandWithName sets the name the command is registered under, overriding
// the default "cloudflared".
func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// NewCommand builds the cloudflared command around the given Cloudflared,
// creating its configured HOME directory. It returns an error if that
// directory cannot be created.
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
						Description: "List forward access",
						Execute:     inst.accessList,
					},
					{
						Name:        "connect",
						Description: "Open access by name",
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
						Name:        "disconnect",
						Description: "Close access by name",
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
				Description: "Manage tunnels",
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
								Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
									fs.Default().Bool("overwrite-dns", false, "Overwrites existing DNS records")
									return nil
								},
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

// Name returns the name the command is registered under.
func (c *Command) Name() string {
	return c.commandTree.Node().Name
}

// Description returns the one-line description shown for the command.
func (c *Command) Description() string {
	return c.commandTree.Node().Description
}

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog shows
// the subcommands; what it cannot show is that the access processes are owned
// by this shell, so the list is empty until something in this session starts
// one.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the symptoms that should reach for it - an unreachable internal service,
// a port that should be listening and is not - because that is how the need
// presents, rather than as "I want to run cloudflared".
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when reaching a private service through a Cloudflare Zero Trust " +
			"tunnel - opening or closing an Access connection so a local port forwards to an " +
			"internal host, checking which tunnels are currently up, or diagnosing a " +
			"connection refused on a port a tunnel should be binding. Also for creating, " +
			"listing or deleting account-level Cloudflare tunnels and their DNS routes.",
	}
}

// Complete returns completion suggestions for the current readline input.
func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

// Validate reports an error if the cloudflared executable is not on PATH.
func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	if _, err := exec.LookPath("cloudflared"); err != nil {
		c.l.Print()
		return errors.New("missing cloudflared executable")
	}

	return nil
}

// Execute runs the cloudflared subcommand selected by the current readline
// input.
func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

// Help returns the help text for the command tree.
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

	return agent.Table(data, agent.WithHeader())
}

func (c *Command) accessConnect(ctx context.Context, r *readline.Readline) error {
	access := c.cloudflared.Config().GetAccesss(r.Args().At(2))
	return c.cloudflared.Connect(ctx, access)
}

func (c *Command) accessDisconnect(ctx context.Context, r *readline.Readline) error {
	access := c.cloudflared.Config().GetAccesss(r.Args().At(2))
	return c.cloudflared.Disconnect(ctx, access)
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

	outDec, err := base64.StdEncoding.DecodeString(string(out))
	if err != nil {
		return err
	}

	c.l.Info(string(outDec))

	return nil
}
