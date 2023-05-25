package squadron

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/slack-go/slack"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/git"
	"github.com/foomo/posh/pkg/util/suggests"
	slackgo "github.com/slack-go/slack"
	"golang.org/x/exp/slices"
)

const All = "all"

type (
	Command struct {
		l              log.Logger
		op             *onepassword.OnePassword
		slack          *slack.Slack
		slackChannelID string
		kubectl        *kubectl.Kubectl
		squadron       *Squadron
		commandTree    tree.Root
		namespaceFn    NamespaceFn
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

func CommandWithSlack(v *slack.Slack) CommandOption {
	return func(o *Command) {
		o.slack = v
	}
}

func CommandWithSlackChannelID(v string) CommandOption {
	return func(o *Command) {
		o.slackChannelID = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, squadron *Squadron, kubectl *kubectl.Kubectl, op *onepassword.OnePassword, opts ...CommandOption) *Command {
	inst := &Command{
		l:              l.Named("squadron"),
		op:             op,
		kubectl:        kubectl,
		squadron:       squadron,
		slackChannelID: "squadron",
		namespaceFn: func(cluster, fleet, squadron string) string {
			if fleet == "default" {
				return squadron
			} else {
				return fmt.Sprintf("%s-%s", fleet, squadron)
			}
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	unitsArg := &tree.Arg{
		Name:     "unit",
		Repeat:   true,
		Optional: true,
		Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
			if value, err := inst.squadron.ListUnits(ctx, r.Args().At(2), r.Args().At(0), r.Args().At(1), true); err != nil {
				return nil
			} else {
				return suggests.List(value)
			}
		},
	}
	slackFlag := func(fs *readline.FlagSets) {
		if inst.slack != nil {
			fs.Internal().Bool("slack", false, "send slack notification")
		}
	}
	profileFlag := func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
		if r.Args().HasIndex(0) {
			fs.Internal().String("profile", "", "Profile to use.")
			if err := fs.Internal().SetValues("profile", inst.kubectl.Cluster(r.Args().At(0)).Profiles(ctx)...); err != nil {
				return err
			}
		}
		return nil
	}
	commonFlags := func(fs *readline.FlagSets) {
		fs.Internal().Bool("no-override", false, "ignore override files")
		fs.Default().Bool("verbose", inst.l.IsLevel(log.LevelDebug), "set verbose level")
		fs.Default().Bool("debug", inst.l.IsLevel(log.LevelTrace), "set debug level")
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

	inst.commandTree = tree.New(&tree.Node{
		Name:        "squadron",
		Description: "Manage your squadron",
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
										return suggests.List(append(value, All))
									}
								},
								Nodes: tree.Nodes{
									{
										Name:        "up",
										Description: "Installs a squadron chart",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											slackFlag(fs)
											commonFlags(fs)
											if err := profileFlag(ctx, r, fs); err != nil {
												return err
											}
											fs.Default().Bool("diff", false, "show diff")
											fs.Default().Bool("push", false, "push image")
											fs.Default().Bool("build", false, "build image")
											fs.Default().Int64("parallel", 0, "number of parallel processes")
											fs.Internal().String("tag", "", "image tag")
											fs.Internal().Bool("create-namespace", false, "create namespace if not exist")
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "list",
										Description: "List squadron units",
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											commonFlags(fs)
											fs.Default().Bool("prefix-squadron", false, "prefix unit names with squadron")
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "down",
										Description: "Uninstalls the squadron chart",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											slackFlag(fs)
											commonFlags(fs)
											if err := profileFlag(ctx, r, fs); err != nil {
												return err
											}
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "push",
										Description: "Push squadron units",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											commonFlags(fs)
											fs.Default().Bool("build", false, "build image")
											fs.Default().Int64("parallel", 0, "number of parallel processes")
											fs.Internal().String("tag", "", "image tag")
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "build",
										Description: "Build or rebuild squadron units",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											commonFlags(fs)
											fs.Default().Bool("push", false, "push image")
											fs.Default().Int64("parallel", 0, "number of parallel processes")
											fs.Internal().String("tag", "", "image tag")
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "status",
										Description: "Display the status of the units",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											commonFlags(fs)
											if err := profileFlag(ctx, r, fs); err != nil {
												return err
											}
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "config",
										Description: "View generated squadron config",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											commonFlags(fs)
											fs.Default().Bool("no-render", false, "push image")
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "rollback",
										Description: "Roll back the squadron chart",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											slackFlag(fs)
											commonFlags(fs)
											fs.Default().String("revision", "", "revision number to rollback to")
											if err := profileFlag(ctx, r, fs); err != nil {
												return err
											}
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "generate",
										Description: "Generate and view the squadron chart",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											commonFlags(fs)
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "template",
										Description: "Render chart templates locally and display the output",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											commonFlags(fs)
											fs.Internal().String("tag", "", "image tag")
											return nil
										},
										Execute: inst.execute,
									},
								},
							},
						},
					},
				},
			},
		},
	})

	return inst
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
	var env []string
	var squadrons []string
	fs := r.FlagSets().Default()
	ifs := r.FlagSets().Internal()
	passFlags := []string{"--"}
	cluster, fleet, squadron, cmd, units := r.Args()[0], r.Args()[1], r.Args()[2], r.Args()[3], r.Args()[4:]

	// retrieve flags
	tag, _ := ifs.GetString("tag")
	noOverride := log.MustGet(ifs.GetBool("no-override"))(c.l)

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	{ // handle 1password
		if c.op != nil {
			if ok, _ := c.op.IsAuthenticated(); !ok {
				c.l.Info("missing 1password session, please login")
				if err := c.op.SignIn(ctx); err != nil {
					return err
				}
			}
		}
	}

	{ // handle pass through flags
		if slices.Contains([]string{"up", "template"}, cmd) {
			passFlags = append(passFlags, "--set", fmt.Sprintf("fleet=%v", fleet))
		}
		if slices.Contains([]string{"up"}, cmd) {
			if log.MustGet(ifs.GetBool("create-namespace"))(c.l) {
				passFlags = append(passFlags, "--create-namespace")
			}
		}
	}

	{ // handle env
		env = append(env, fmt.Sprintf("FLEET=%s", fleet))
		if tag != "" {
			env = append(env, fmt.Sprintf("TAG=%q", tag))
		}
		env = append(env, c.kubectl.Cluster(cluster).Env(profile))
	}

	{ // handle squadrons
		if squadron == All {
			if value, err := c.squadron.List(); err == nil {
				squadrons = value
			}
		} else {
			squadrons = []string{squadron}
		}
	}

	{ // handle slack
		if ok, _ := ifs.GetBool("slack"); ok {
			if err := c.notify(ctx, cmd, cluster, fleet, squadron, tag, units); err != nil {
				return err
			}
		}
	}

	for _, s := range squadrons {
		env := append(env, fmt.Sprintf("SQUADRON=%s", s))
		flags := r.FlagSets().Default().Args()
		files := strings.Join(c.squadron.GetFiles(s, cluster, fleet, !noOverride), ",")
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
			Args(fs.Visited().Args()...).
			Args(r.AdditionalFlags()...).
			Args(passFlags...).
			Args(r.AdditionalArgs()...).
			Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) notify(ctx context.Context, cmd, cluster, fleet, squadron, tag string, units []string) error {
	if tag == "" {
		tag = "latest"
	}

	user, err := git.ConfigUserName(ctx, c.l)
	if err != nil {
		c.l.Debug("failed to get git user:", err.Error())
		user = "unknown"
	}

	ref, err := git.Ref(ctx, c.l)
	if err != nil {
		c.l.Debug("failed to get git ref:", err.Error())
		ref = "unknown"
	}

	var msg *slackgo.SectionBlock

	switch cmd {
	case "up":
		if squadron == All {
			msg = c.slack.MarkdownSection(fmt.Sprintf("🚢 Full deployment to *%s* | *%s* _(%s)_", cluster, fleet, tag))
		} else if len(units) == 0 {
			msg = c.slack.MarkdownSection(fmt.Sprintf("🛥 Deployment to *%s*\n\n- %s.all | *%s* _(%s)_\n", cluster, squadron, fleet, tag))
		} else {
			str := make([]string, 0, len(units))
			for _, unit := range units {
				str = append(str, "- "+squadron+"."+unit)
			}
			msg = c.slack.MarkdownSection(fmt.Sprintf("🛶 Deployment to *%s* | *%s* _(%s)_\n\n%s\n", cluster, fleet, tag, strings.Join(str, "\n")))
		}
	case "down":
		if squadron == All {
			msg = c.slack.MarkdownSection(fmt.Sprintf("🪦 Full uninstallation of *%s* | *%s*", cluster, fleet))
		} else if len(units) == 0 {
			msg = c.slack.MarkdownSection(fmt.Sprintf("💀 Uninstalling from *%s*\n\n- %s.all | *%s*\n", cluster, squadron, fleet))
		} else {
			str := make([]string, 0, len(units))
			for _, unit := range units {
				str = append(str, "- "+squadron+"."+unit)
			}
			msg = c.slack.MarkdownSection(fmt.Sprintf("🗑 Uninstalling from *%s* | *%s*\n\n%s\n", cluster, fleet, strings.Join(str, "\n")))
		}
	case "rollback":
		if squadron == "all" {
			msg = c.slack.MarkdownSection(fmt.Sprintf("⏬ Full roll back of *%s* | *%s*", cluster, fleet))
		} else if len(units) == 0 {
			msg = c.slack.MarkdownSection(fmt.Sprintf("⏪ Rollback in *%s*\n\n- %s.all | *%s*\n", cluster, squadron, fleet))
		} else {
			str := make([]string, 0, len(units))
			for _, unit := range units {
				str = append(str, "- "+squadron+"."+unit)
			}
			msg = c.slack.MarkdownSection(fmt.Sprintf("🔙 Rollback in *%s* | *%s*\n\n%s\n", cluster, fleet, strings.Join(str, "\n")))
		}
	default:
		c.l.Debug("skipping notification for cmd:", cmd)
	}

	blockOpt := slackgo.MsgOptionBlocks(
		msg,
		slackgo.NewContextBlock("", slackgo.NewTextBlockObject("mrkdwn", ref+" by "+user, false, false)),
		c.slack.DividerSection(),
	)
	fallbackOpt := slackgo.MsgOptionText(fmt.Sprintf("Deployment to %s | %s", cluster, fleet), false)

	return c.slack.Send(
		ctx,
		c.slack.Channel(c.slackChannelID),
		slackgo.MsgOptionCompose(fallbackOpt, blockOpt),
	)
}
