package squadron

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"golang.org/x/exp/slices"
)

type (
	Command struct {
		l           log.Logger
		op          *onepassword.OnePassword
		kubectl     *kubectl.Kubectl
		squadron    *Squadron
		commandTree *tree.Root
		namespaceFn NamespaceFn
	}
	NamespaceFn   func(cluster, fleet, squadron string) string
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithNamespaceFn(v NamespaceFn) CommandOption {
	return func(o *Command) {
		o.namespaceFn = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, squadron *Squadron, kubectl *kubectl.Kubectl, op *onepassword.OnePassword, opts ...CommandOption) *Command {
	inst := &Command{
		l:        l.Named("squadron"),
		op:       op,
		kubectl:  kubectl,
		squadron: squadron,
		namespaceFn: func(cluster, fleet, squadron string) string {
			if fleet == "default" {
				return squadron
			} else {
				return fmt.Sprintf("%s-%s", fleet, squadron)
			}
		},
	}

	unitsArg := &tree.Arg{
		Name:     "unit",
		Repeat:   true,
		Optional: true,
		Suggest: func(ctx context.Context, t *tree.Root, r *readline.Readline) []goprompt.Suggest {
			if value, err := inst.squadron.ListUnits(ctx, r.Args().At(2), r.Args().At(0), r.Args().At(1), true); err != nil {
				return nil
			} else {
				return suggests.List(value)
			}
		},
	}
	commonFlags := func(fs *readline.FlagSet) {
		fs.Bool("no-override", false, "ignore override files")
		fs.Bool("verbose", inst.l.IsLevel(log.LevelDebug), "set verbose level")
		fs.Bool("debug", inst.l.IsLevel(log.LevelTrace), "set debug level")
	}

	clusterValues := func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
		var ret []string
		for _, cluster := range inst.kubectl.Clusters() {
			if _, ok := inst.squadron.cfg.Cluster(cluster.Name()); ok {
				ret = append(ret, cluster.Name())
			}
		}
		return suggests.List(ret)
	}

	inst.commandTree = &tree.Root{
		Name:        "squadron",
		Description: "manage your squadron",
		Nodes: tree.Nodes{
			{
				Name:   "cluster",
				Values: clusterValues,
				Nodes: tree.Nodes{
					{
						Name: "fleet",
						Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
							var ret []string
							if cluster, ok := inst.squadron.cfg.Cluster(r.Args().At(0)); ok {
								ret = cluster.Fleets
							}
							return suggests.List(ret)
						},
						Nodes: tree.Nodes{
							{
								Name: "squadron",
								Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
									if value, err := inst.squadron.List(); err != nil {
										inst.l.Debug(err.Error())
										return nil
									} else {
										return suggests.List(append(value, "all"))
									}
								},
								Nodes: tree.Nodes{
									{
										Name:        "up",
										Description: "installs a squadron chart",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											commonFlags(fs)
											fs.Bool("diff", false, "show diff")
											fs.Bool("push", false, "push image")
											fs.Bool("build", false, "build image")
											fs.String("tag", "", "image tag")
											fs.Int64("parallel", 0, "number of parallel processes")
											return nil
										},
										PassThroughFlags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											fs.Bool("create-namespace", false, "create namespace if not exist")
											return nil
										},
										Execute: inst.up,
									},
									{
										Name:        "list",
										Description: "list squadron units",
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											commonFlags(fs)
											fs.Bool("prefix-squadron", false, "prefix unit names with squadron")
											return nil
										},
										Execute: inst.list,
									},
									{
										Name:        "down",
										Description: "uninstalls the squadron chart",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											commonFlags(fs)
											return nil
										},
										Execute: inst.down,
									},
									{
										Name:        "push",
										Description: "push squadron units",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											commonFlags(fs)
											fs.Bool("build", false, "build image")
											return nil
										},
										Execute: inst.push,
									},
									{
										Name:        "build",
										Description: "build or rebuild squadron units",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											commonFlags(fs)
											fs.Bool("push", false, "push image")
											fs.String("tag", "", "image tag")
											fs.Int64("parallel", 0, "number of parallel processes")
											return nil
										},
										Execute: inst.build,
									},
									{
										Name:        "status",
										Description: "display the status of the units",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											commonFlags(fs)
											return nil
										},
										Execute: inst.status,
									},
									{
										Name:        "config",
										Description: "view generated squadron config",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											commonFlags(fs)
											fs.Bool("no-render", false, "push image")
											return nil
										},
										Execute: inst.config,
									},
									{
										Name:        "rollback",
										Description: "roll back the squadron chart",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											commonFlags(fs)
											fs.String("revision", "", "revision number to rollback to")
											return nil
										},
										Execute: inst.rollback,
									},
									{
										Name:        "generate",
										Description: "generate and view the squadron chart",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											commonFlags(fs)
											return nil
										},
										Execute: inst.generate,
									},
									{
										Name:        "template",
										Description: "render chart templates locally and display the output",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSet) error {
											commonFlags(fs)
											return nil
										},
										Execute: inst.template,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return inst
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c *Command) Name() string {
	return c.commandTree.Name
}

func (c *Command) Description() string {
	return c.commandTree.Description
}

func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help() string {
	return `Manage your squadron.

Usage:
  squadron [cluster] [fleet] [squadron] [cmd]

Available Commands:
  up <fleet...>
  down <fleet...>

Examples:
  k9s example-cluster my-namespace
`
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) up(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r)
}

func (c *Command) list(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r)
}

func (c *Command) down(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r)
}

func (c *Command) push(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r)
}

func (c *Command) build(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r)
}

func (c *Command) status(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r)
}

func (c *Command) config(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r)
}

func (c *Command) rollback(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r)
}

func (c *Command) generate(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r)
}

func (c *Command) template(ctx context.Context, r *readline.Readline) error {
	return c.execute(ctx, r)
}

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	flags := r.Flags()
	cluster, fleet, squadron, cmd, units := r.Args()[0], r.Args()[1], r.Args()[2], r.Args()[3], r.Args()[4:]

	if c.op != nil {
		if ok, _ := c.op.Session(); !ok {
			c.l.Info("missing 1password session, please login")
			if err := c.op.SignIn(ctx); err != nil {
				return err
			}
		}
	}

	passFlags := r.PassThroughFlags()
	if slices.Contains([]string{"up", "template"}, cmd) {
		if len(passFlags) == 0 {
			passFlags = append(passFlags, "--")
		}
		passFlags = append(passFlags, "--set", fmt.Sprintf("fleet=%v", fleet))
	}

	var env []string
	env = append(env, fmt.Sprintf("FLEET=%s", fleet))
	if value := r.FlagSet().GetString("tag"); value != "" {
		env = append(env, fmt.Sprintf("TAG=%q", value))
		flags = flags.Splice(flags.IndexOf("--tag"), 2)
	}
	env = append(env, c.kubectl.Cluster(cluster).Env())

	var squadrons []string
	if squadron == "all" {
		if value, err := c.squadron.List(); err == nil {
			squadrons = value
		}
	} else {
		squadrons = []string{squadron}
	}

	for _, s := range squadrons {
		env := append(env, fmt.Sprintf("SQUADRON=%s", s))
		flags := flags
		files := strings.Join(c.squadron.GetFiles(s, cluster, fleet, !r.FlagSet().GetBool("no-override")), ",")
		if slices.Contains([]string{"up", "down", "rollback", "status", "template"}, cmd) {
			flags = append(flags, "--namespace", c.namespaceFn(cluster, fleet, s))
		}

		c.l.Infof("Fleet:    %s", fleet)
		c.l.Infof("Squadron: %s", s)

		if err := shell.New(ctx, c.l, "squadron", cmd, "--file", files).
			Env(env...).
			Dir(path.Join(c.squadron.cfg.Path, s)).
			Args(units...).
			Args(flags...).
			Args(passFlags...).
			Args(r.AdditionalArgs()...).
			Run(); err != nil {
			return err
		}
	}
	return nil
}
