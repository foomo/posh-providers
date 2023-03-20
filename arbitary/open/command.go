package open

import (
	"context"
	"net/url"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/browser"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

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
		Description: "Open an external url",
		Args: tree.Args{
			{
				Name: "router",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					var ret []goprompt.Suggest
					for s, router := range inst.cfg {
						ret = append(ret, goprompt.Suggest{Text: s, Description: router.Description})
					}
					return ret
				},
			},
			{
				Name: "route",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					var ret []goprompt.Suggest
					if value, ok := inst.cfg[r.Args().At(0)]; ok {
						for s, route := range value.Routes {
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
	case r.Args().LenGt(2):
		return errors.New("too many arguments")
	}
	if _, ok := c.cfg[r.Args().At(0)]; !ok {
		return errors.New("invalid [router] argument")
	} else if _, ok := c.cfg[r.Args().At(0)].Routes[r.Args().At(1)]; !ok {
		return errors.New("invalid [rout] argument")
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
	router := c.cfg[r.Args().At(0)]
	route := router.Routes[r.Args().At(1)]

	u, err := url.Parse(router.URL + route.Path)
	if err != nil {
		return err
	}

	if route.Username != nil && route.Password != nil {
		if username, err := c.op.Get(ctx, *route.Username); err != nil {
			return err
		} else if password, err := c.op.Get(ctx, *route.Password); err != nil {
			return err
		} else {
			u.User = url.UserPassword(username, password)
		}
	}

	return browser.OpenURL(u)
}
