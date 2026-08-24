package gcx

import (
	"context"
	_ "embed"
	"errors"
	"strings"

	exec2 "github.com/foomo/posh-providers/pkg/exec"
	"github.com/foomo/posh-providers/pkg/os"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/spf13/viper"
)

//go:embed SKILL.md
var skill string

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

type (
	Command struct {
		l           log.Logger
		name        string
		cfg         Config
		cache       cache.Namespace
		configKey   string
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

func CommandWithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("gcx"),
		name:      "gcx",
		configKey: "gcx",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	inst.cache = cache.Get(inst.name)

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run gcx against a configured Grafana server",
		Args: tree.Args{
			{
				Name:        "env",
				Description: "Environment name from the gcx config, supplying the server, org and token",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.cfg.EnvNames())
				},
			},
			{
				Name:        "command",
				Description: "Command and flags passed through to gcx, completed by gcx itself",
				Repeat:      true,
				Optional:    true,
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return exec2.Complete(ctx, inst.name, 1, r)
				},
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
// shows only the env name, not that the second argument is the whole gcx CLI,
// that the env silently scopes which Grafana server destructive verbs hit, that
// every call resolves a live credential via 1Password or an executed tokenCmd,
// or that flags after a `--` separator are dropped.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names Grafana resources rather than the gcx binary, since a request arrives as
// "export the dashboards", not as the name of the CLI that does it.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when managing a Grafana server's contents from the CLI - listing, " +
			"exporting, importing, migrating or deleting dashboards, datasources, folders or " +
			"alert rules - against one of this project's configured Grafana environments. " +
			"Also for gcx itself, or when a Grafana call fails on a missing token.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	envName := r.Args().At(0)

	env, err := c.cfg.Env(envName)
	if err != nil {
		return err
	}

	var token string

	switch {
	case len(env.Token) > 0:
		token, err = os.ExpandEnv(ctx, env.Token)
		if err != nil {
			return err
		}
	case len(env.TokenCmd) > 0:
		v, err := shell.New(ctx, c.l).Args(env.TokenCmd...).Output()
		if err != nil {
			return err
		}

		token = strings.TrimSpace(string(v))
	default:
		return errors.New("no token provided")
	}

	envs := []string{
		"DO_NOT_TRACK=1",
		"GRAFANA_ORG_ID=" + env.OrgID,
		"GCX_TELEMETRY=0",
		"GCX_NO_UPDATE_NOTIFIER=1",
		"GRAFANA_SERVER=" + env.URL,
		"GRAFANA_TOKEN=" + token,
	}

	return shell.New(ctx, c.l, "gcx").
		Args(r.Args().From(1)...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Env(envs...).
		Env(env.Env...).
		Run()
}
