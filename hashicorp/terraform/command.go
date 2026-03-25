package terraform

import (
	"context"
	"errors"
	"os/exec"

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
		cfg         Config
		name        string
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

func WithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("terraform"),
		name:      "terraform",
		configKey: "terraform",
		cache:     cache.Get("terraform"),
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
		Description: "Run terraform commands",
		Nodes: tree.Nodes{
			{
				Name:        "workspace",
				Values:      inst.getWorkspaces,
				Description: "Workspace to operate on",
				Nodes: tree.Nodes{
					{
						Name:        "init",
						Description: "Prepare your working directory for other commands",
						Flags:       inst.addAuthFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "validate",
						Description: "Check whether the configuration is valid",
						Flags:       inst.addAuthFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "plan",
						Description: "Show changes required by the current configuration",
						Args: tree.Args{
							{
								Name:        "target",
								Description: "Limit to a specific module or resource (optional)",
								Optional:    true,
								Repeat:      true,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.cfg.WorkspaceTargets(r.Args().At(0)))
								},
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							if err := inst.addAuthFlags(ctx, r, fs); err != nil {
								return err
							}

							fs.Default().Bool("refresh-only", false, "Select the 'refresh only' planning mode")
							fs.Default().String("var-file", "", "Load variable values from a file")

							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "apply",
						Description: "Create or update infrastructure",
						Args: tree.Args{
							{
								Name:        "target",
								Description: "Limit to a specific module or resource (optional)",
								Optional:    true,
								Repeat:      true,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.cfg.WorkspaceTargets(r.Args().At(0)))
								},
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							if err := inst.addAuthFlags(ctx, r, fs); err != nil {
								return err
							}

							fs.Default().Bool("auto-approve", false, "Skip interactive approval")
							fs.Default().Bool("refresh-only", false, "Select the 'refresh only' planning mode")
							fs.Default().String("var-file", "", "Load variable values from a file")

							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "destroy",
						Description: "Destroy previously-created infrastructure",
						Args: tree.Args{
							{
								Name:        "target",
								Description: "Limit to a specific module or resource (optional)",
								Optional:    true,
								Repeat:      true,
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return suggests.List(inst.cfg.WorkspaceTargets(r.Args().At(0)))
								},
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							if err := inst.addAuthFlags(ctx, r, fs); err != nil {
								return err
							}

							fs.Default().Bool("auto-approve", false, "Skip interactive approval")
							fs.Default().String("var-file", "", "Load variable values from a file")

							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "refresh",
						Description: "Update the state to match remote systems",
						Flags:       inst.addAuthFlags,
						Execute:     inst.execute,
					},
					{
						Name:        "output",
						Description: "Show output values from your root module",
						Args: tree.Args{
							{
								Name:        "name",
								Description: "Output name (optional)",
								Optional:    true,
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							if err := inst.addAuthFlags(ctx, r, fs); err != nil {
								return err
							}

							fs.Default().Bool("raw", false, "Print the raw string directly")
							fs.Default().Bool("json", false, "Output in JSON format")

							return nil
						},
						Execute: inst.execute,
					},
					{
						Name:        "import",
						Description: "Associate existing infrastructure with a Terraform resource",
						Args: tree.Args{
							{
								Name:        "address",
								Description: "Resource address to import into",
							},
							{
								Name:        "id",
								Description: "Resource ID to import",
							},
						},
						Flags:   inst.addAuthFlags,
						Execute: inst.execute,
					},
					{
						Name:        "state",
						Description: "Advanced state management",
						Nodes: tree.Nodes{
							{
								Name:        "list",
								Description: "List resources in the state",
								Flags:       inst.addAuthFlags,
								Execute:     inst.executeState,
							},
							{
								Name:        "show",
								Description: "Show a resource in the state",
								Args: tree.Args{
									{
										Name:        "address",
										Description: "Resource address to show",
									},
								},
								Flags:   inst.addAuthFlags,
								Execute: inst.executeState,
							},
							{
								Name:        "mv",
								Description: "Move an item in the state",
								Args: tree.Args{
									{
										Name:        "source",
										Description: "Source address",
									},
									{
										Name:        "destination",
										Description: "Destination address",
									},
								},
								Flags:   inst.addAuthFlags,
								Execute: inst.executeState,
							},
							{
								Name:        "rm",
								Description: "Remove instances from the state",
								Args: tree.Args{
									{
										Name:        "address",
										Description: "Resource address to remove",
									},
								},
								Flags:   inst.addAuthFlags,
								Execute: inst.executeState,
							},
						},
					},
					{
						Name:        "unlock",
						Description: "Unlock a stuck lock on the current workspace",
						Args: tree.Args{
							{
								Name:        "lockId",
								Description: "Terraform state lock ID",
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							if err := inst.addAuthFlags(ctx, r, fs); err != nil {
								return err
							}

							fs.Default().Bool("force", false, "Don't ask for input confirmation")

							return nil
						},
						Execute: inst.unlock,
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
	if _, err := exec.LookPath("terraform"); err != nil {
		c.l.Print()
		return errors.New("missing terraform executable")
	}

	return nil
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) addAuthFlags(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
	fs.Internal().String("service-principal", "", "Service principal to use for authentication")
	return fs.Internal().SetValues("service-principal", c.cfg.ServicePrincipalNames()...)
}

func (c *Command) authEnv(r *readline.Readline, workspace string) ([]string, error) {
	sp, err := r.FlagSets().Internal().GetString("service-principal")
	if err != nil {
		return nil, err
	}

	if sp != "" {
		spc, err := c.cfg.ServicePrincipal(sp)
		if err != nil {
			return nil, err
		}

		return []string{
			"ARM_TENANT_ID=" + spc.TenantID,
			"ARM_CLIENT_ID=" + spc.ClientID,
			"ARM_CLIENT_SECRET=" + spc.ClientSecret,
			"ARM_SUBSCRIPTION_ID=" + spc.SubscriptionID,
		}, nil
	}

	// User account (az CLI) auth — inject subscription ID if configured
	if sub, err := c.cfg.Subscription(workspace); err == nil {
		return []string{"ARM_SUBSCRIPTION_ID=" + sub.ID}, nil
	}

	return nil, nil
}

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	workspace := r.Args().At(0)
	command := r.Args().At(1)

	env, err := c.authEnv(r, workspace)
	if err != nil {
		return err
	}

	cmd := shell.New(ctx, c.l, "terraform", command).
		Dir(c.cfg.WorkspacePath(workspace)).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Env(env...)

	switch command {
	case "apply", "plan", "destroy":
		// Positional args are module/resource targets selected from autocomplete;
		// they must be passed as -target= flags, not as bare positional args.
		for _, target := range r.Args().From(2) {
			cmd = cmd.Args("-target=" + target)
		}
	case "init":
		cmd = cmd.Args(c.cfg.BackendInitArgs(workspace)...)
	default:
		cmd = cmd.Args(r.Args().From(2)...)
	}

	return cmd.Run()
}

func (c *Command) executeState(ctx context.Context, r *readline.Readline) error {
	workspace := r.Args().At(0)
	subcommand := r.Args().At(2)

	env, err := c.authEnv(r, workspace)
	if err != nil {
		return err
	}

	return shell.New(ctx, c.l, "terraform", "state", subcommand).
		Dir(c.cfg.WorkspacePath(workspace)).
		Args(r.Args().From(3)...).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Env(env...).
		Run()
}

func (c *Command) unlock(ctx context.Context, r *readline.Readline) error {
	workspace := r.Args().At(0)
	lockID := r.Args().At(2)

	env, err := c.authEnv(r, workspace)
	if err != nil {
		return err
	}

	return shell.New(ctx, c.l, "terraform", "force-unlock").
		Dir(c.cfg.WorkspacePath(workspace)).
		Args(r.FlagSets().Default().Visited().Args()...).
		Args(lockID).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Env(env...).
		Run()
}

func (c *Command) getWorkspaces(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.cache.GetSuggests("workspaces", func() any {
		workspaces := c.cfg.WorkspaceNames()
		if len(workspaces) == 0 {
			c.l.Debug("no terraform workspaces found")
		}

		return suggests.List(workspaces)
	})
}
