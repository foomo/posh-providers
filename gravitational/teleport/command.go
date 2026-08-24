package teleport

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
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
		cache       cache.Cache
		kubectl     *kubectl.Kubectl
		teleport    *Teleport
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

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, teleport *Teleport, kubectl *kubectl.Kubectl, opts ...CommandOption) *Command {
	inst := &Command{
		l:        l.Named("teleport"),
		name:     "teleport",
		cache:    cache,
		kubectl:  kubectl,
		teleport: teleport,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Manage access points through teleport; logs in when called without a subcommand",
		Execute:     inst.auth,
		Nodes: tree.Nodes{
			{
				Name:        "auth",
				Description: "Log in to a cluster and retrieve the session certificate",
				Execute:     inst.auth,
			},
			{
				Name:        "kubeconfig",
				Description: "Retrieve credentials to access remote cluster.",
				Args: tree.Args{
					{
						Name:        "cluster",
						Description: "Name of the cluster.",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.teleport.Clusters(ctx))
						},
					},
				},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("profile", "", "Profile to use.")
					return fs.Internal().SetValues("profile", "teleport")
				},
				Execute: inst.kubeconfig,
			},
			{
				Name:        "database",
				Description: "Retrieve credentials to access remote database.",
				Args: tree.Args{
					{
						Name:        "database",
						Description: "Name of the database.",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.teleport.Databases(ctx))
						},
					},
				},
				Execute: inst.database,
			},
			{
				Name:        "app",
				Description: "Retrieve credentials to access remote app.",
				Args: tree.Args{
					{
						Name:        "name",
						Description: "Name of the app",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.teleport.Apps(ctx))
						},
					},
				},
				Execute: inst.app,
			},
			{
				Name:        "logout",
				Description: "Log out",
				Execute:     inst.logout,
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog lists
// five verbs and cannot show that all of them need an interactive browser SSO
// an agent cannot complete; that the bare root logs in rather than printing
// help; that `kubeconfig` replaces the profile's existing config, restoring it
// if the login fails; that nothing completes until authenticated, and then only
// what the configured labels match; or that an ambiguous cluster alias is
// rejected rather than resolved.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the access symptoms - no kubeconfig, expired certificate - because an
// agent reaches for this when something else it wanted to use is unreachable.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when getting access to a cluster, database or app that sits behind " +
			"Teleport - logging in via SSO, writing a kubeconfig for a remote cluster, " +
			"obtaining database or app credentials, or logging out. Also when kubectl reports " +
			"no context or an expired certificate, or a tsh session needs renewing.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) app(ctx context.Context, r *readline.Readline) error {
	app := r.Args().At(1)
	appArgs := c.teleport.Config().Apps[app]

	return shell.New(ctx, c.l, "tsh", "apps", "login").
		Args(appArgs...).
		Args(app).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) database(ctx context.Context, r *readline.Readline) error {
	databse := r.Args().At(1)

	return shell.New(ctx, c.l, "tsh", "db", "login",
		"--db-user", c.teleport.Config().Database.EnvUser(),
		databse,
	).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}

func (c *Command) kubeconfig(ctx context.Context, r *readline.Readline) error {
	ifs := r.FlagSets().Internal()
	cluster := c.kubectl.Cluster(r.Args().At(1))

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	clusterName, err := c.teleport.cfg.Kubernetes.Name(cluster.Name())
	if err != nil {
		return err
	}

	// `tsh kube login` merges into whatever is already at KUBECONFIG (upstream's
	// kubeconfig.Update loads the path first), so the old file has to go to stop
	// stale contexts accumulating. Move it aside rather than deleting it: a failed
	// login - expired session, no network - would otherwise leave the profile
	// with no kubeconfig at all, having destroyed a working one.
	stash, err := stashFile(cluster.Config(profile))
	if err != nil {
		return err
	}

	// generate & filter new config
	if err := shell.New(ctx, c.l, "tsh", "kube", "login", clusterName).
		Env(cluster.Env(profile)).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run(); err != nil {
		return errors.Join(err, stash.Restore())
	}

	return stash.Discard()
}

// stash holds a file moved aside so a failed operation can put it back.
type stash struct {
	path string
	// copy is empty when the path did not exist, in which case Restore only has
	// to remove whatever was written in its place.
	copy string
}

// stashFile moves path out of the way if it exists. Call Restore to put it back
// on failure, or Discard once the operation has succeeded.
func stashFile(path string) (*stash, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return &stash{path: path}, nil
	} else if err != nil {
		return nil, err
	}

	copyPath := path + ".posh-stash"
	if err := os.Rename(path, copyPath); err != nil {
		return nil, err
	}

	return &stash{path: path, copy: copyPath}, nil
}

// Restore puts the original file back, discarding anything written in its place.
func (s *stash) Restore() error {
	if err := ignoreNotExist(os.Remove(s.path)); err != nil {
		return err
	}

	if s.copy == "" {
		return nil
	}

	return os.Rename(s.copy, s.path)
}

// Discard drops the preserved copy, keeping whatever now sits at the path.
func (s *stash) Discard() error {
	if s.copy == "" {
		return nil
	}

	return ignoreNotExist(os.Remove(s.copy))
}

func ignoreNotExist(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

func (c *Command) auth(ctx context.Context, r *readline.Readline) error {
	if err := shell.New(ctx, c.l, "tsh", "login",
		fmt.Sprintf("--proxy=%s", c.teleport.Config().Hostname),
		"--auth=github",
	).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run(); err != nil {
		return err
	}

	return nil
}

func (c *Command) logout(ctx context.Context, r *readline.Readline) error {
	if err := shell.New(ctx, c.l, "tsh", "logout").
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run(); err != nil {
		return err
	}

	return nil
}
