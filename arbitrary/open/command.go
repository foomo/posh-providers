package open

import (
	"context"
	_ "embed"
	"net/url"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/browser"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

//go:embed SKILL.md
var skill string

type (
	Command struct {
		l           log.Logger
		op          *onepassword.OnePassword
		cfg         Config
		name        string
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

func WithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, op *onepassword.OnePassword, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("open"),
		op:        op,
		name:      "open",
		configKey: "open",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Open a configured url in the browser",
		Args: tree.Args{
			{
				Name:        "router",
				Description: "Router name from the open config, supplying the base url",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					var ret []goprompt.Suggest
					for s, router := range inst.cfg {
						ret = append(ret, goprompt.Suggest{Text: s, Description: router.Description})
					}

					return ret
				},
			},
			{
				Name:        "route",
				Description: "Route name under the router; repeat to descend nested routes",
				Repeat:      true,
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					var ret []goprompt.Suggest

					if value, ok := inst.cfg[r.Args().At(0)]; ok {
						for s, route := range value.RoutesForPath(r.Args().From(1)) {
							ret = append(ret, goprompt.Suggest{Text: s, Description: route.Description})
						}
					}

					return ret
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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	switch {
	case r.Args().LenIs(0):
		return errors.New("missing [router] argument")
	case r.Args().LenIs(1):
		return errors.New("missing [route] argument")
	}

	if router, ok := c.cfg[r.Args().At(0)]; !ok {
		return errors.New("invalid [router] argument")
	} else if route := router.RouteForPath(r.Args().From(1)); route.Path == "" {
		return errors.New("invalid [route] argument")
	}

	return nil
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
// cannot show that a route may embed 1Password credentials in the opened URL,
// that the router silently decides which environment is reached, or that each
// argument after the router descends one level of the configured route tree.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	router := c.cfg[r.Args().At(0)]
	route := router.RouteForPath(r.Args().From(1))

	u, err := url.Parse(router.URL + route.Path)
	if err != nil {
		return err
	}

	if route.BasicAuth != nil {
		var (
			username string
			password string
		)
		{
			secret := *route.BasicAuth

			secret.Field = "username"
			if value, err := c.op.Get(ctx, secret); err != nil {
				return err
			} else {
				username = value
			}
		}
		{
			secret := *route.BasicAuth

			secret.Field = "password"
			if value, err := c.op.Get(ctx, secret); err != nil {
				return err
			} else {
				password = value
			}
		}

		u.User = url.UserPassword(username, password)
	}

	return browser.OpenURL(ctx, u)
}
