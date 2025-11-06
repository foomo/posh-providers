package pulumi

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/foomo/posh-providers/google/gcloud"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
		name        string
		gcloud      *gcloud.GCloud
		op          *onepassword.OnePassword
		cfg         Config
		cache       cache.Namespace
		configKey   string
		commandTree tree.Root
	}
	NamespaceFn   func(cluster, fleet, squadron string) string
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

func NewCommand(l log.Logger, gcloud *gcloud.GCloud, op *onepassword.OnePassword, cache cache.Cache, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		name:      "pulumi",
		configKey: "pulumi",
		op:        op,
		gcloud:    gcloud,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.l = l.Named(inst.name)
	inst.cache = cache.Get(inst.name)

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	if err := os.Setenv("PULUMI_HOME", env.Path(inst.cfg.ConfigPath)); err != nil {
		return nil, err
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "pulumi",
		Description: "Open the pulumi dashboard",
		Nodes: tree.Nodes{
			{
				Name:        "env",
				Values:      inst.completeEnvs,
				Description: "Name of the environment",
				Nodes: tree.Nodes{
					{
						Name:        "backend",
						Description: "Manage state backends",
						Nodes: tree.Nodes{
							{
								Name:        "create",
								Description: "Create a new object storage backend",
								Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
									fs.Default().String("debug", "", "Show full logs")
									fs.Default().String("tags", "", "Quoted string with space-separated tags")
									fs.Default().String("vebose", "", "Increase logging verbosity")

									return nil
								},
								Execute: func(ctx context.Context, r *readline.Readline) error {
									be, err := inst.cfg.Backend(r.Args().At(0))
									if err != nil {
										return err
									}

									inst.l.Info("creating storage bucket:", be.Bucket)

									return shell.New(ctx, inst.l, "gcloud", "storage", "buckets", "create", fmt.Sprintf("gs://%s", be.Bucket)).
										Args("--location", be.Location).
										Args("--project", be.Project).
										Run()
								},
							},
							{
								Name:        "login",
								Description: "Log into your object storage backend",
								Execute: func(ctx context.Context, r *readline.Readline) error {
									// be, _, err := inst.backendKey(ctx, r.Args().At(0))
									be, err := inst.cfg.Backend(r.Args().At(0))
									if err != nil {
										return err
									}
									// Configure Google ADC auth
									inst.l.Info("configuring google ADC auth")

									if err := shell.New(ctx, inst.l, "gcloud", "auth", "application-default", "login").
										Args("--project", be.Project).
										Run(); err != nil {
										return err
									}

									return shell.New(ctx, inst.l, "pulumi", "login", fmt.Sprintf("gs://%s", be.Bucket)).
										// Env("GOOGLE_PROJECT=" + be.Project).
										Env("GOOGLE_APPLICATION_CREDENTIALS=" + env.Path("devops/config/gcloud/application_default_credentials.json")).
										Run()
								},
							},
						},
					},
					{
						Name:        "stack",
						Description: "Manage stacks and view stack state",
						Args: tree.Args{
							{
								Name:    "project",
								Suggest: inst.completeProjects,
							},
							{
								Name:    "stack",
								Suggest: inst.completeStacks,
							},
							{
								Name: "command",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return []goprompt.Suggest{
										{Text: "init", Description: "Create an empty stack with the given name, ready for updates"},
										{Text: "output", Description: "Show a stack's output properties"},
										{Text: "history", Description: "Display history for a stack"},
									}
								},
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("help", false, "Show command help")
							fs.Default().Int("verbose", 3, "Enable verbose logging")

							return nil
						},
						Execute: inst.executeStack,
					},
					{
						Name:        "up",
						Description: "Create or update the resources in a stack",
						Args: tree.Args{
							{
								Name:    "project",
								Suggest: inst.completeProjects,
							},
							{
								Name:    "stack",
								Suggest: inst.completeStacks,
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("debug", false, "Print detailed debugging output during resource operations")
							fs.Default().Bool("diff", false, "Display operation as a rich diff showing the overall change")
							fs.Default().Bool("expect-no-changes", false, "Return an error if any changes occur during this update")
							fs.Default().Bool("help", false, "Show command help")
							fs.Default().Bool("target-dependents", false, "Allows updating of dependent targets discovered but not specified in --target list")
							fs.Default().Int("verbose", 3, "Enable verbose logging")
							fs.Default().StringArray("target", nil, "Specify a single resource URN to update")
							fs.Default().StringArray("target-replace", nil, "Specify a single resource URN to replace")

							return nil
						},
						Execute: inst.executeStack,
					},
					{
						Name:        "destroy",
						Description: "Destroy all existing resources in the stack",
						Args: tree.Args{
							{
								Name:    "project",
								Suggest: inst.completeProjects,
							},
							{
								Name:    "stack",
								Suggest: inst.completeStacks,
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("debug", false, "Print detailed debugging output during resource operations")
							fs.Default().Bool("diff", false, "Display operation as a rich diff showing the overall change")
							fs.Default().Bool("exclude-protected", false, "Do not destroy protected resources")
							fs.Default().Bool("help", false, "Show command help")
							fs.Default().Bool("remove", false, "Remove the stack and its config file after all resources in the stack have been deleted")
							fs.Default().Bool("target-dependents", false, "Allows updating of dependent targets discovered but not specified in --target list")
							fs.Default().Int("verbose", 3, "Enable verbose logging")
							fs.Default().StringArray("target", nil, "Specify a single resource URN to update")

							return nil
						},
						Execute: inst.executeStack,
					},
					{
						Name:        "preview",
						Description: "Show a preview of updates to a stack's resources",
						Args: tree.Args{
							{
								Name:    "project",
								Suggest: inst.completeProjects,
							},
							{
								Name:    "stack",
								Suggest: inst.completeStacks,
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("debug", false, "Print detailed debugging output during resource operations")
							fs.Default().Bool("diff", false, "Display operation as a rich diff showing the overall change")
							fs.Default().Bool("expect-no-changes", false, "Return an error if any changes occur during this update")
							fs.Default().Bool("help", false, "Show command help")
							fs.Default().Bool("target-dependents", false, "Allows updating of dependent targets discovered but not specified in --target list")
							fs.Default().Int("verbose", 3, "Enable verbose logging")
							fs.Default().StringArray("target", nil, "Specify a single resource URN to update")
							fs.Default().StringArray("target-replace", nil, "Specify a single resource URN to replace")

							return nil
						},
						Execute: inst.executeStack,
					},
					{
						Name:        "cancel",
						Description: "Cancel a stack's currently running update, if any",
						Args: tree.Args{
							{
								Name:    "project",
								Suggest: inst.completeProjects,
							},
							{
								Name:    "stack",
								Suggest: inst.completeStacks,
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("help", false, "Show command help")
							return nil
						},
						Execute: inst.executeStack,
					},
					{
						Name:        "refresh",
						Description: "Refresh the resources in a stack",
						Args: tree.Args{
							{
								Name:    "project",
								Suggest: inst.completeProjects,
							},
							{
								Name:    "stack",
								Suggest: inst.completeStacks,
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("clear-pending-creates", false, "Clear all pending creates, dropping them from the state")
							fs.Default().Bool("debug", false, "Print detailed debugging output during resource operations")
							fs.Default().Bool("diff", false, "Display operation as a rich diff showing the overall change")
							fs.Default().Bool("expect-no-changes", false, "Return an error if any changes occur during this update")
							fs.Default().Bool("help", false, "Show command help")
							fs.Default().Bool("show-replacement-steps", false, "Show detailed resource replacement creates and deletes instead of a single step")
							fs.Default().Bool("show-sames", false, "Show resources that needn't be updated because they haven't changed, alongside those that d")
							fs.Default().StringArray("import-pending-creates", nil, "A list of form [[URN ID]...] describing the provider IDs of pending creates")
							fs.Default().StringArray("target", nil, "Specify a single resource URN to update")

							return nil
						},
						Execute: inst.executeStack,
					},
					{
						Name:        "state",
						Description: "Edit the current stack's state",
						Args: tree.Args{
							{
								Name:    "project",
								Suggest: inst.completeProjects,
							},
							{
								Name:    "stack",
								Suggest: inst.completeStacks,
							},
							{
								Name: "command",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return []goprompt.Suggest{
										{Text: "delete", Description: "Deletes a resource from a stack's state"},
										{Text: "rename", Description: "Renames a resource from a stack's state"},
										{Text: "unprotect", Description: "Unprotect resources in a stack's state"},
										{Text: "upgrade", Description: "Migrates the current backend to the latest supported version"},
									}
								},
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("help", false, "Show command help")
							return nil
						},
						Execute: inst.executeStack,
					},
					{
						Name:        "import",
						Description: "Import resources into an existing stack",
						Args: tree.Args{
							{
								Name:    "project",
								Suggest: inst.completeProjects,
							},
							{
								Name:    "stack",
								Suggest: inst.completeStacks,
							},
							{
								Name: "type",
							},
							{
								Name: "name",
							},
							{
								Name: "id",
							},
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("debug", false, "Print detailed debugging output during resource operations")
							fs.Default().Bool("diff", false, "Display operation as a rich diff showing the overall change")
							fs.Default().Bool("help", false, "Show command help")
							fs.Default().String("file", "", "The path to a JSON-encoded file containing a list of resources to import")
							fs.Default().String("from", "", "Invoke a converter to import the resources")
							fs.Default().String("out", "", "The path to the file that will contain the generated resource declarations")
							fs.Default().String("parent", "", "The name and URN of the parent resource in the format name=urn")
							fs.Default().StringArray("properties", nil, "The property names to use for the import in the format name1,name")

							return nil
						},
						Execute: inst.executeStack,
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

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) completeEnvs(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	//nolint:forcetypeassert
	return c.cache.Get("envs", func() any {
		entries, err := os.ReadDir(c.cfg.Path)
		if err != nil {
			c.l.Debug(err.Error())
			return []goprompt.Suggest{}
		}

		var ret []string

		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				ret = append(ret, e.Name())
			}
		}

		return suggests.List(ret)
	}).([]goprompt.Suggest)
}

func (c *Command) configureStack(ctx context.Context, stackEnv, proj, stack string, be Backend, passphrase string) error {
	filename := path.Join(c.cfg.Path, stackEnv, proj, fmt.Sprintf("Pulumi.%s.op", stack))

	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	out, err := shell.New(ctx, c.l, "cat", filename, "|", "op", "inject").Output()
	if err != nil {
		return errors.Wrap(err, "failed to inject onepassword")
	}

	var args []string

	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && strings.Contains(line, "=") {
			args = append(args, "--secret", line)
		}
	}

	if len(args) == 0 {
		return nil
	}

	return shell.New(ctx, c.l, "pulumi", "config", "set-all").
		Args(args...).
		Args("--stack", stack).
		Dir(path.Join(c.cfg.Path, stackEnv, proj)).
		Env("PULUMI_BACKEND_URL=" + fmt.Sprintf("gs://%s", be.Bucket)).
		Env("PULUMI_CONFIG_PASSPHRASE=" + passphrase).
		Env("GOOGLE_PROJECT=" + be.Project).
		Env("GOOGLE_APPLICATION_CREDENTIALS=" + env.Path("devops/config/gcloud/application_default_credentials.json")).
		Args().
		Run()
}

func (c *Command) executeStack(ctx context.Context, r *readline.Readline) error {
	e := r.Args().At(0)
	proj := r.Args().At(2)
	stack := r.Args().At(3)

	be, err := c.cfg.Backend(e)
	if err != nil {
		return err
	}

	passphrase, err := c.op.Get(ctx, be.Passphrase)
	if err != nil {
		return err
	}

	if err := c.configureStack(ctx, e, proj, stack, be, passphrase); err != nil {
		return err
	}

	return shell.New(ctx, c.l, "pulumi", r.Args().At(1)).
		Args("--stack", stack).
		Args(r.Args().From(4)...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Env("PULUMI_BACKEND_URL=" + fmt.Sprintf("gs://%s", be.Bucket)).
		Env("PULUMI_CONFIG_PASSPHRASE=" + passphrase).
		Env("GOOGLE_PROJECT=" + be.Project).
		Env("GOOGLE_APPLICATION_CREDENTIALS=" + env.Path("devops/config/gcloud/application_default_credentials.json")).
		Dir(path.Join(c.cfg.Path, e, proj)).
		Run()
}

func (c *Command) completeProjects(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	e := r.Args().At(0)
	//nolint:forcetypeassert
	return c.cache.Get("projects-"+e, func() any {
		entries, err := os.ReadDir(path.Join(c.cfg.Path, e))
		if err != nil {
			c.l.Debug(err.Error())
			return []goprompt.Suggest{}
		}

		var ret []string

		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				ret = append(ret, e.Name())
			}
		}

		return suggests.List(ret)
	}).([]goprompt.Suggest)
}

func (c *Command) completeStacks(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	e := r.Args().At(0)
	project := r.Args().At(2)
	//nolint:forcetypeassert
	return c.cache.Get("stacks-"+e+"-"+project, func() any {
		entries, err := os.ReadDir(path.Join(c.cfg.Path, e, project))
		if err != nil {
			c.l.Debug(err.Error())
			return []goprompt.Suggest{}
		}

		var ret []string

		for _, e := range entries {
			if !e.IsDir() && len(e.Name()) > 11 && strings.HasPrefix(e.Name(), "Pulumi.") && strings.HasSuffix(e.Name(), ".yaml") {
				ret = append(ret, strings.TrimSuffix(strings.TrimPrefix(e.Name(), "Pulumi."), ".yaml"))
			}
		}

		return suggests.List(ret)
	}).([]goprompt.Suggest)
}
