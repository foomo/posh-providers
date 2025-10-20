package terragrunt

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type (
	Command struct {
		l           log.Logger
		op          *onepassword.OnePassword
		cfg         Config
		name        string
		cache       cache.Namespace
		configKey   string
		commandTree tree.Root
	}
	StackNameProvider func(path string) string
	CommandOption     func(*Command)
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

func NewCommand(l log.Logger, op *onepassword.OnePassword, cache cache.Cache, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("terragrunt"),
		op:        op,
		name:      "terragrunt",
		cache:     cache.Get("terragrunt"),
		configKey: "terragrunt",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	if err := os.Setenv("TERRAGRUNT_DOWNLOAD", env.Path(inst.cfg.CachePath)); err != nil {
		return nil, err
	}

	stackArgs := tree.Args{
		{
			Name:        "stacks",
			Description: "Stacks to run",
			Repeat:      true,
			Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
				return inst.getStacks(ctx, r)
			},
		},
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run terragrunt commands",
		Nodes: tree.Nodes{
			{
				Name:        "env",
				Values:      inst.getEnvs,
				Description: "Environment to provision",
				Nodes: tree.Nodes{
					{
						Name:        "site",
						Values:      inst.getSites,
						Description: "Site to provision",
						Nodes: tree.Nodes{
							{
								Name:        "secrets",
								Description: "Render secret templates",
								Execute:     inst.secrets,
							},
							// terraform: main commands
							{
								Name:        "init",
								Description: "Prepare your working directory for other commands",
								Args:        stackArgs,
								Execute:     inst.execute,
							},
							{
								Name:        "validate",
								Description: "Check whether the configuration is valid",
								Args:        stackArgs,
								Execute:     inst.execute,
							},
							{
								Name:        "plan",
								Description: "Show changes required by the current configuration",
								Args:        stackArgs,
								Execute:     inst.execute,
							},
							{
								Name:        "apply",
								Description: "Create or update infrastructure",
								Args:        stackArgs,
								Execute:     inst.execute,
							},
							{
								Name:        "destroy",
								Description: "Destroy previously-created infrastructure",
								Args:        stackArgs,
								Execute:     inst.execute,
							},
							{
								Name:        "refresh",
								Description: "Update the state to match remote systems",
								Args:        stackArgs,
								Execute:     inst.execute,
							},
							{
								Name:        "output",
								Description: "Show output values from your root module",
								Args:        stackArgs,
								Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
									fs.Default().String("raw", "", "Print the raw string directly")
									return nil
								},
								Execute: inst.execute,
							},
							// {
							// 	Name: "command",
							// 	Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
							// 		return []prompt.Suggest{
							// 			// terraform: main commands
							// 			{Text: "init", Description: "Prepare your working directory for other commands"},
							// 			{Text: "validate", Description: "Check whether the configuration is valid"},
							// 			{Text: "plan", Description: "Show changes required by the current configuration"},
							// 			{Text: "apply", Description: "Create or update infrastructure"},
							// 			{Text: "destroy", Description: "Destroy previously-created infrastructure"},
							// 			// terraform: other commands
							// 			{Text: "hclfmt", Description: "Reformat your configuration in the standard style"},
							// 			// {Text: "console", Description: "Try Terragrunt expressions at an interactive command prompt"},
							// 			{Text: "force-unlock", Description: "Unlock a stuck lock on the current workspace"},
							// 			{Text: "get", Description: "Install or upgrade remote Terragrunt modules"},
							// 			{Text: "graph", Description: "Generate a Graphviz graph of the steps in an operation"},
							// 			{Text: "import", Description: "Associate existing infrastructure with a Terragrunt resource"},
							// 			// {Text: "login", Description: "Obtain and save credentials for a remote host"},
							// 			// {Text: "logout", Description: "Remove locally-stored credentials for a remote host"},
							// 			{Text: "output", Description: "Show output values from your root module"},
							// 			{Text: "providers:lock", Description: "Write out dependency locks for the configured providers"},
							// 			{Text: "refresh", Description: "Update the state to match remote systems"},
							// 			{Text: "show", Description: "Show the current state or a saved plan"},
							// 			{Text: "state", Description: "Advanced state management"},
							// 			// {"taint", "Mark a resource instance as not fully functional"},
							// 			// {"untaint", "Remove the 'tainted' state from a resource instance"},
							// 			// {"version", "Show the current Terragrunt version"},
							// 			// {"workspace", "Workspace management"},
							// 		}
							// 	},
							// 	Description: "Terragrunt command to execute",
							// 	Args: tree.Args{
							// 		{
							// 			Name:        "stacks",
							// 			Description: "Stacks to run",
							// 			Repeat:      true,
							// 			Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							// 				return inst.getStacks(ctx, r)
							// 			},
							// 		},
							// 	},
							// 	Execute: inst.execute,
							// },
						},
						Execute: inst.execute,
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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	if _, err := exec.LookPath("terragrunt"); err != nil {
		c.l.Print()
		return errors.New("missing terragrunt executable")
	}

	return nil
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) secrets(ctx context.Context, r *readline.Readline) error {
	envName := r.Args().At(0)
	siteName := r.Args().At(1)

	// validate stack & change dir
	values, err := files.Find(ctx, c.cfg.StacksPath(envName, siteName), "secrets.tpl.yaml", files.FindWithIsFile(true))
	if err != nil {
		return err
	}

	c.l.Info("Rendering secret templates...")

	for _, value := range values {
		c.l.Info("└  " + value)

		if err := c.op.RenderFileTo(ctx, value, strings.Replace(value, ".tpl.yaml", ".yaml", 1)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	envName := r.Args().At(0)
	siteName := r.Args().At(1)
	command := r.Args().At(2)
	stacks := r.Args().From(3)

	c.l.Info("Running terragrunt...")

	for _, stack := range stacks {
		c.l.Info("└  " + stack)

		if err := shell.New(ctx, c.l, "terragrunt", command).
			Args(r.AdditionalFlags()...).
			Dir(path.Join(c.cfg.StacksPath(envName, siteName), stack)).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) getEnvs(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.cache.GetSuggests("envs", func() any {
		return suggests.List(c.cfg.EnvNames())
	})
}

func (c *Command) getSites(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	envName := r.Args().At(0)

	return c.cache.GetSuggests("sites-"+envName, func() any {
		return suggests.List(c.cfg.SiteNames(envName))
	})
}

func (c *Command) getStacks(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	envName := r.Args().At(0)
	site := r.Args().At(1)

	return c.cache.GetSuggests("stacks-"+envName+"-"+site, func() any {
		stacks, err := c.cfg.StackNames(ctx, envName, site)
		if err != nil {
			c.l.Debug("failed to retrieve stacks", zap.Error(err))
		}

		return suggests.List(stacks)
	})
}
