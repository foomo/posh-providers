package sesamy

import (
	"bytes"
	"context"
	_ "embed"
	"sort"
	"strings"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
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
		op          *onepassword.OnePassword
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

func CommandWithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, op *onepassword.OnePassword, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("sesamy"),
		op:        op,
		name:      "sesamy",
		configKey: "sesamy",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.config); err != nil {
		return nil, err
	}

	// if err := os.Setenv("SESAMY_SCOPE", inst.name); err != nil {
	// 	return nil, err
	// }

	configArg := &tree.Arg{
		Name:        "config",
		Description: "Config set name from the sesamy config; every set if omitted",
		Optional:    true,
		Suggest:     inst.completePaths,
	}
	flags := func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
		fs.Default().Bool("verbose", false, "show verbose output")
		return nil
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Run sesamy against a named config set",
		Nodes: tree.Nodes{
			{
				Name:        "config",
				Description: "Print the merged config with secrets resolved",
				Args:        tree.Args{configArg},
				Execute:     inst.conf,
			},
			{
				Name:        "tags",
				Description: "List available tags",
				Args:        tree.Args{configArg},
				Flags:       flags,
				Execute:     inst.tags,
			},
			{
				Name:        "typescript",
				Description: "Generate typescript definitions",
				Args:        tree.Args{configArg},
				Flags:       flags,
				Execute:     inst.typescript,
			},
			{
				Name:        "open",
				Description: "Open links in the browser",
				Args: tree.Args{
					{
						Name:        "name",
						Description: "Which console to open",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return []goprompt.Suggest{
								{Text: "ga"},
								{Text: "gtm-web"},
								{Text: "gtm-server"},
							}
						},
					},
					configArg,
				},
				Execute: inst.open,
			},
			{
				Name:        "provision",
				Description: "Provision Google Tag Manager",
				Nodes: tree.Nodes{
					{
						Name:        "all",
						Description: "Provision Web & Server Container",
						Args:        tree.Args{configArg},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().StringSlice("tags", nil, "list of tags to run")
							return flags(ctx, r, fs)
						},
						Execute: inst.provisionAll,
					},
					{
						Name:        "web",
						Description: "Provision Web Container",
						Args:        tree.Args{configArg},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().StringSlice("tags", nil, "list of tags to run")
							return flags(ctx, r, fs)
						},
						Execute: inst.provisionWeb,
					},
					{
						Name:        "server",
						Description: "Provision Server Container",
						Args:        tree.Args{configArg},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().StringSlice("tags", nil, "list of tags to run")
							return flags(ctx, r, fs)
						},
						Execute: inst.provisionServer,
					},
				},
			},
			{
				Name:        "list",
				Description: "List Google Tag Manager Resources",
				Nodes: tree.Nodes{
					{
						Name:        "web",
						Description: "List Web Container Resources",
						Args: tree.Args{
							{
								Name:        "resource",
								Description: "Name of the resource",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return []goprompt.Suggest{
										{Text: "built-in-variables"},
										{Text: "environments"},
										{Text: "folders"},
										{Text: "gtag-config"},
										{Text: "tags"},
										{Text: "templates"},
										{Text: "templates-data"},
										{Text: "transformations"},
										{Text: "triggers"},
										{Text: "variables"},
										{Text: "workspaces"},
										{Text: "zones"},
									}
								},
							},
							configArg,
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("raw", false, "print raw output")
							return nil
						},
						Execute: inst.listWeb,
					},
					{
						Name:        "server",
						Description: "List Server Container Resources",
						Args: tree.Args{
							{
								Name:        "resource",
								Description: "Name of the resource",
								Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
									return []goprompt.Suggest{
										{Text: "built-in-variables"},
										{Text: "environments"},
										{Text: "clients"},
										{Text: "folders"},
										{Text: "gtag-config"},
										{Text: "status"},
										{Text: "tags"},
										{Text: "templates"},
										{Text: "templates-data"},
										{Text: "transformations"},
										{Text: "triggers"},
										{Text: "variables"},
										{Text: "workspaces"},
										{Text: "zones"},
									}
								},
							},
							configArg,
						},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("raw", false, "print raw output")
							return nil
						},
						Execute: inst.listServer,
					},
				},
			},
			{
				Name:        "diff",
				Description: "Print Google Tag Manager Status diff",
				Nodes: tree.Nodes{
					{
						Name:        "web",
						Description: "Print Web Container Status diff",
						Args:        tree.Args{configArg},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("raw", false, "print raw output")
							return nil
						},
						Execute: inst.diffWeb,
					},
					{
						Name:        "server",
						Description: "Print Server Container Status diff",
						Args:        tree.Args{configArg},
						Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
							fs.Default().Bool("raw", false, "print raw output")
							return nil
						},
						Execute: inst.diffServer,
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The rendered tree
// cannot show that an omitted config argument runs the verb against every
// configured set, that `provision` writes to live GTM containers, or that every
// verb renders 1Password secrets into the config it pipes to sesamy.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// leads with Google Tag Manager and analytics tracking rather than the tool
// name, because that is the vocabulary a request for this work arrives in.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when working with this project's Google Tag Manager or Google " +
			"Analytics setup - provisioning tags, triggers or variables into a web or " +
			"server container, diffing configured against live GTM state, listing " +
			"container resources, generating the TypeScript tracking definitions, or " +
			"opening the GA/GTM console. Also for sesamy config sets.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) merge(ctx context.Context, path string) ([]byte, error) {
	filenames, ok := c.config[path]
	if !ok {
		return nil, errors.New("invalid config key: " + path)
	}

	var (
		conf = koanf.Conf{Delim: "/"}
		k    = koanf.NewWithConf(conf)
	)

	for _, filename := range filenames {
		if err := k.Load(file.Provider(filename), yaml.Parser()); err != nil {
			return nil, errors.Wrap(err, "error loading config file: "+filename)
		}
	}

	return k.Marshal(yaml.Parser())
}

func (c *Command) conf(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Printing sesamy config ...")

	for _, value := range paths {
		c.l.Info("└ " + value)

		b, err := c.merge(ctx, value)
		if err != nil {
			return errors.Wrap(err, "failed to merge config")
		}

		out, err := c.op.Render(ctx, string(b))
		if err != nil {
			pterm.Error.Println(string(b))
			return errors.Wrap(err, "failed to render secrets")
		}

		if err := shell.New(ctx, c.l, "sesamy", "config").
			Args(r.Flags()...).
			Args("--config", "-").
			Args(r.AdditionalArgs()...).
			Stdin(bytes.NewReader(out)).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) tags(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy tags ...")

	for _, value := range paths {
		c.l.Info("└ " + value)

		b, err := c.merge(ctx, value)
		if err != nil {
			return errors.Wrap(err, "failed to merge config")
		}

		out, err := c.op.Render(ctx, string(b))
		if err != nil {
			pterm.Error.Println(string(b))
			return errors.Wrap(err, "failed to render secrets")
		}

		if err := shell.New(ctx, c.l, "sesamy", "tags").
			Args(r.Flags()...).
			Args("--config", "-").
			Args(r.AdditionalArgs()...).
			Stdin(bytes.NewReader(out)).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) open(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy open ...")

	for _, value := range paths {
		c.l.Info("└ " + value)

		b, err := c.merge(ctx, value)
		if err != nil {
			return errors.Wrap(err, "failed to merge config")
		}

		out, err := c.op.Render(ctx, string(b))
		if err != nil {
			pterm.Error.Println(string(b))
			return errors.Wrap(err, "failed to render secrets")
		}

		if err := shell.New(ctx, c.l, "sesamy", "open", r.Args().At(1)).
			Args(r.Flags()...).
			Args("--config", "-").
			Args(r.AdditionalArgs()...).
			Stdin(bytes.NewReader(out)).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) typescript(ctx context.Context, r *readline.Readline) error {
	var paths []string
	if r.Args().HasIndex(1) {
		paths = []string{r.Args().At(1)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy typescript ...")

	for _, value := range paths {
		c.l.Info("└ " + value)

		b, err := c.merge(ctx, value)
		if err != nil {
			return errors.Wrap(err, "failed to merge config")
		}

		out, err := c.op.Render(ctx, string(b))
		if err != nil {
			pterm.Error.Println(string(b))
			return errors.Wrap(err, "failed to render secrets")
		}

		if err := shell.New(ctx, c.l, "sesamy", "typescript").
			Args(r.Flags()...).
			Args("--config", "-").
			Args(r.AdditionalArgs()...).
			Stdin(bytes.NewReader(out)).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) provision(ctx context.Context, r *readline.Readline, cmd string) error {
	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy provision " + cmd + " ...")

	for _, value := range paths {
		c.l.Info("└ " + value)

		b, err := c.merge(ctx, value)
		if err != nil {
			return errors.Wrap(err, "failed to merge config")
		}

		out, err := c.op.Render(ctx, string(b))
		if err != nil {
			pterm.Error.Println(string(b))
			return errors.Wrap(err, "failed to render secrets")
		}

		if err := shell.New(ctx, c.l, "sesamy", "provision", cmd).
			Args(r.Flags()...).
			Args("--config", "-").
			Args(r.AdditionalArgs()...).
			Stdin(bytes.NewReader(out)).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) provisionAll(ctx context.Context, r *readline.Readline) error {
	if err := c.provision(ctx, r, "web"); err != nil {
		return err
	}

	if err := c.provision(ctx, r, "server"); err != nil {
		return err
	}

	return nil
}

func (c *Command) provisionWeb(ctx context.Context, r *readline.Readline) error {
	return c.provision(ctx, r, "web")
}

func (c *Command) provisionServer(ctx context.Context, r *readline.Readline) error {
	return c.provision(ctx, r, "server")
}

func (c *Command) listWeb(ctx context.Context, r *readline.Readline) error {
	return c.list(ctx, r, "web")
}

func (c *Command) listServer(ctx context.Context, r *readline.Readline) error {
	return c.list(ctx, r, "server")
}

func (c *Command) list(ctx context.Context, r *readline.Readline, cmd string) error {
	fs := r.FlagSets().Default()
	resource := r.Args().At(2)

	var paths []string
	if r.Args().HasIndex(3) {
		paths = []string{r.Args().At(3)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy list " + cmd + " ...")

	for _, value := range paths {
		c.l.Info("└ " + value)

		b, err := c.merge(ctx, value)
		if err != nil {
			return errors.Wrap(err, "failed to merge config")
		}

		out, err := c.op.Render(ctx, string(b))
		if err != nil {
			pterm.Error.Println(string(b))
			return errors.Wrap(err, "failed to render secrets")
		}

		if err := shell.New(ctx, c.l, "sesamy", "list", cmd, resource).
			Args(fs.Visited().Args()...).
			Args("--config", "-").
			Args(r.AdditionalArgs()...).
			Stdin(bytes.NewReader(out)).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) diffWeb(ctx context.Context, r *readline.Readline) error {
	return c.diff(ctx, r, "web")
}

func (c *Command) diffServer(ctx context.Context, r *readline.Readline) error {
	return c.diff(ctx, r, "server")
}

func (c *Command) diff(ctx context.Context, r *readline.Readline, cmd string) error {
	fs := r.FlagSets().Default()

	var paths []string
	if r.Args().HasIndex(2) {
		paths = []string{r.Args().At(2)}
	} else {
		paths = c.paths(ctx)
	}

	c.l.Info("Running sesamy diff " + cmd + " ...")

	for _, value := range paths {
		c.l.Info("└ " + value)

		b, err := c.merge(ctx, value)
		if err != nil {
			return errors.Wrap(err, "failed to merge config")
		}

		out, err := c.op.Render(ctx, string(b))
		if err != nil {
			pterm.Error.Println(string(b))
			return errors.Wrap(err, "failed to render secrets")
		}

		if err := shell.New(ctx, c.l, "sesamy", "diff", cmd).
			Args(fs.Visited().Args()...).
			Args("--config", "-").
			Args(r.AdditionalArgs()...).
			Stdin(bytes.NewReader(out)).
			Run(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) completePaths(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
	return suggests.List(c.paths(ctx))
}

func (c *Command) paths(ctx context.Context) []string {
	keys := lo.Keys(c.config)
	sort.Strings(keys)

	return keys
}
