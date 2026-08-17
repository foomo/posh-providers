package gcx

import (
	"context"
	"errors"
	"strings"

	exec2 "github.com/foomo/posh-providers/pkg/exec"
	"github.com/foomo/posh-providers/pkg/os"
	"github.com/foomo/posh/pkg/cache"
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
		Description: "Run gcx commands",
		Args: tree.Args{
			{
				Name:        "env",
				Description: "Environment name",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.cfg.EnvNames())
				},
			},
			{
				Name:        "command",
				Description: "GCX command",
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
