package bruno

import (
	"context"
	"os"
	"path"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	gap "github.com/muesli/go-app-paths"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
		op          *onepassword.OnePassword
		name        string
		appName     string
		config      Config
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

func CommandWithAppName(v string) CommandOption {
	return func(o *Command) {
		o.appName = v
	}
}

func CommandWithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

func CommandWithOnePassword(v *onepassword.OnePassword) CommandOption {
	return func(o *Command) {
		o.op = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("bruno"),
		name:      "bruno",
		appName:   "Bruno",
		configKey: "bruno",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.config); err != nil {
		return nil, err
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run Bruno requests",
		Nodes: tree.Nodes{
			{
				Name:        "list",
				Description: "List available requests",
				Execute:     inst.list,
			},
			{
				Name:        "env",
				Description: "Render secrets env",
				Execute:     inst.env,
			},
			{
				Name:        "run",
				Description: "Run the Bruno cli",
				Args: tree.Args{
					{
						Name:        "env",
						Description: "Environment name",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.config.Environments())
						},
					},
					{
						Name:        "request",
						Description: "Request to run",
						Repeat:      true,
						Optional:    true,
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.config.Requests())
						},
					},
				},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Default().String("format", "", "Environment variables")
					fs.Default().String("env", "", "Overwrite a single environment variable")
					fs.Default().String("env-var", "", "Overwrite a single environment variable")
					fs.Default().Bool("bail", false, "Stop execution after a failure of a request, test, or assertion")
					fs.Default().Bool("insecure", false, "Allow insecure server connections")
					fs.Default().Bool("tests-only", false, "Only run requests that have tests")
					fs.Default().Bool("verbose", false, "Allow verbose output for debugging purpose")

					return nil
				},
				Execute: inst.run,
			},
			{
				Name:        "open",
				Description: "Open the Bruno app",
				Execute:     inst.open,
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

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) list(ctx context.Context, r *readline.Readline) error {
	t := pterm.TreeNode{
		Text: c.config.Filename() + ":",
	}
	{
		child := pterm.TreeNode{
			Text: "Environments:",
		}
		for _, request := range c.config.Environments() {
			child.Children = append(child.Children, pterm.TreeNode{
				Text: request,
			})
		}

		if len(child.Children) > 0 {
			t.Children = append(t.Children, child)
		}
	}
	{
		child := pterm.TreeNode{
			Text: "Requests:",
		}
		for _, request := range c.config.Requests() {
			child.Children = append(child.Children, pterm.TreeNode{
				Text: request,
			})
		}

		if len(child.Children) > 0 {
			t.Children = append(t.Children, child)
		}
	}

	return pterm.DefaultTree.WithRoot(t).Render()
}

func (c *Command) run(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()

	args := []string{
		"--env", r.Args().At(1),
	}
	if r.Args().LenGte(3) {
		args = append(args, r.Args().From(2)...)
	}

	return shell.New(ctx, c.l, "bru", "run").
		Args(args...).
		Args(fs.Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Dir(c.config.Filename()).
		Run()
}

func (c *Command) env(ctx context.Context, r *readline.Readline) error {
	if c.op == nil {
		return errors.New("you must provide a one-password configuration")
	}

	envFilename := path.Join(c.config.Filename(), ".env")

	templateFilename := path.Join(c.config.Filename(), "bruno.env")
	if _, err := os.Stat(templateFilename); err != nil {
		return err
	}

	c.l.Info("rendering secrets file:", envFilename)

	return shell.New(ctx, c.l, "op", "inject", "-f", "-i", templateFilename, "-o", envFilename).Quiet().Run()
}

func (c *Command) open(ctx context.Context, r *readline.Readline) error {
	prefFilename, err := gap.NewScope(gap.User, c.name).DataPath("preferences.json")
	if err != nil {
		return err
	}

	pref, err := NewPreferences(prefFilename)
	if err != nil {
		return err
	}

	if err := pref.AddLastOpenedCollection(c.config.Filename()); err != nil {
		return err
	}

	if err := pref.Save(prefFilename); err != nil {
		return err
	}

	return shell.New(ctx, c.l, "open", "-a", c.appName).Run()
}
