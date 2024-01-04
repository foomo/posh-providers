package squadron

import (
	"context"
	"fmt"
	"strings"

	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/slack-go/slack"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/git"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
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
		cache          cache.Namespace
		kubectl        *kubectl.Kubectl
		squadron       *Squadron
		commandTree    tree.Root
		namespaceFn    NamespaceFn
	}
	NamespaceFn   func(cluster, fleet string) string
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

func NewCommand(l log.Logger, squadron *Squadron, kubectl *kubectl.Kubectl, op *onepassword.OnePassword, cache cache.Cache, opts ...CommandOption) *Command {
	inst := &Command{
		l:              l.Named("squadron"),
		op:             op,
		cache:          cache.Get("squadron"),
		kubectl:        kubectl,
		squadron:       squadron,
		slackChannelID: "squadron",
		namespaceFn: func(cluster, fleet string) string {
			if fleet == "default" {
				return "{{ .Squadron }}"
			} else {
				return fmt.Sprintf("%s-{{ .Squadron }}", fleet)
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
			cluster := r.Args().At(0)
			fleet := r.Args().At(1)
			squadron := r.Args().At(2)
			return inst.cache.Get(fmt.Sprintf("units-%s-%s-%s", squadron, cluster, fleet), func() any {
				if value, err := inst.squadron.ListUnits(ctx, squadron, cluster, fleet, true); err != nil {
					return []goprompt.Suggest{}
				} else {
					return suggests.List(value)
				}
			}).([]goprompt.Suggest)
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
										Name:        "build",
										Description: "Build or rebuild squadron units",
										Args:        tree.Args{unitsArg},
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											commonFlags(fs)
											fs.Default().Bool("push", false, "push image")
											fs.Default().Int64("parallel", 1, "number of parallel processes")
											fs.Default().StringSlice("tags", nil, "list of tags to include or exclude")
											// build args
											fs.Get("build-args").StringArray("build-arg", nil, "set build-time variables")
											fs.Internal().String("tag", "", "image tag")
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
											fs.Default().Bool("no-render", false, "don't render the config template")
											fs.Default().StringSlice("tags", nil, "list of tags to include or exclude")
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "diff",
										Description: "Shows the diff between the installed and local chart",
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											commonFlags(fs)
											fs.Default().Int64("parallel", 1, "number of parallel processes")
											fs.Default().StringSlice("tags", nil, "list of tags to include or exclude")
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
											fs.Default().Int64("parallel", 1, "number of parallel processes")
											fs.Default().StringSlice("tags", nil, "list of tags to include or exclude")
											if err := profileFlag(ctx, r, fs); err != nil {
												return err
											}
											return nil
										},
										Execute: inst.execute,
									},
									{
										Name:        "list",
										Description: "List squadron units",
										Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
											commonFlags(fs)
											fs.Default().Bool("charts", false, "include charts")
											fs.Default().Bool("builds", false, "include builds")
											fs.Default().StringSlice("tags", nil, "list of tags to include or exclude")
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
											fs.Default().Int64("parallel", 1, "number of parallel processes")
											fs.Default().StringSlice("tags", nil, "list of tags to include or exclude")
											// build args
											fs.Get("build-args").StringArray("build-arg", nil, "set build-time variables")
											fs.Internal().String("tag", "", "image tag")
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
											fs.Default().Int64("parallel", 1, "number of parallel processes")
											fs.Default().String("revision", "", "revision number to rollback to")
											fs.Default().StringSlice("tags", nil, "list of tags to include or exclude")
											if err := profileFlag(ctx, r, fs); err != nil {
												return err
											}
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
											fs.Default().Int64("parallel", 1, "number of parallel processes")
											fs.Default().StringSlice("tags", nil, "list of tags to include or exclude")
											if err := profileFlag(ctx, r, fs); err != nil {
												return err
											}
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
											fs.Default().Int64("parallel", 1, "number of parallel processes")
											fs.Default().StringSlice("tags", nil, "list of tags to include or exclude")
											fs.Internal().String("tag", "", "image tag")
											return nil
										},
										Execute: inst.execute,
									},
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
											fs.Default().Bool("push", false, "push image")
											fs.Default().Bool("build", false, "build image")
											fs.Default().Int64("parallel", 1, "number of parallel processes")
											fs.Default().StringSlice("tags", nil, "list of tags to include or exclude")
											// build args
											fs.Get("build-args").StringArray("build-arg", nil, "set build-time variables")
											// internal
											fs.Internal().String("tag", "", "image tag")
											fs.Internal().Bool("create-namespace", false, "create namespace if not exist")
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
	var buildArgs []string
	fs := r.FlagSets().Default()
	ifs := r.FlagSets().Internal()
	bfs := r.FlagSets().Get("build-args")
	passFlags := []string{"--"}
	cluster, fleet, squadron, cmd, units := r.Args()[0], r.Args()[1], r.Args()[2], r.Args()[3], r.Args()[4:]

	// retrieve flags
	tag, _ := ifs.GetString("tag")
	noOverride := log.MustGet(ifs.GetBool("no-override"))(c.l)

	// try retrieve profile
	profile, _ := ifs.GetString("profile")

	{ // handle 1password
		if c.op != nil {
			if ok, _ := c.op.IsAuthenticated(ctx); !ok {
				c.l.Info("missing 1password session, please login")
				if err := c.op.SignIn(ctx); err != nil {
					return err
				}
			}
		}
	}

	{ // handle pass through flags
		if slices.Contains([]string{"up", "diff", "template"}, cmd) {
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

	{ // handle slack
		if ok, _ := ifs.GetBool("slack"); ok {
			if err := c.notify(ctx, cmd, cluster, fleet, squadron, tag, units); err != nil {
				return err
			}
		}
	}

	c.l.Infof("Fleet:    %s", fleet)
	c.l.Infof("Cluster:  %s", cluster)

	if value, err := bfs.GetStringArray("build-arg"); err == nil && len(value) > 0 {
		for _, v := range value {
			buildArgs = append(buildArgs, "--build-arg", v)
		}
	}
	// flatten build args
	if len(buildArgs) > 0 {
		buildArgs = []string{fmt.Sprintf("--build-args='%s'", strings.Join(buildArgs, " "))}
	}

	sh := shell.New(ctx, c.l, "squadron", cmd).
		Args("--file", strings.Join(c.squadron.GetFiles("", cluster, fleet, !noOverride), ",")).
		Dir(c.squadron.cfg.Path).
		Env(env...)

	if squadron != All {
		sh.Args(squadron).Args(units...)
	}

	flags := r.FlagSets().Default().Args()
	if slices.Contains([]string{"up", "diff", "down", "rollback", "status", "template"}, cmd) {
		flags = append(flags, "--namespace", c.namespaceFn(cluster, fleet))
	}

	if err := sh.
		Args(flags...).
		Args(fs.Visited().Args()...).
		Args(buildArgs...).
		Args(passFlags...).
		Args(r.AdditionalArgs()...).
		Run(); err != nil {
		return errors.Wrap(err, "failed to execute squadron")
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
