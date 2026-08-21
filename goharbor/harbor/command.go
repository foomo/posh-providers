package harbor

import (
	"context"
	_ "embed"
	"os"
	"strings"

	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util"
	"github.com/foomo/posh/pkg/util/browser"
	"github.com/google/go-github/v47/github"
	"github.com/pterm/pterm"
	"golang.org/x/oauth2"
)

//go:embed SKILL.md
var skill string

type (
	Command struct {
		l           log.Logger
		name        string
		harbor      *Harbor
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

func NewCommand(l log.Logger, harbor *Harbor, opts ...CommandOption) *Command {
	inst := &Command{
		l:      l.Named("harbor"),
		name:   "harbor",
		harbor: harbor,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Sign in to the configured Harbor registry",
		Execute:     inst.auth,
		Nodes: tree.Nodes{
			{
				Name:        "auth",
				Description: "Open the Harbor login page in a browser",
				Execute:     inst.auth,
			},
			{
				Name:        "docker",
				Description: "Log the local docker daemon in to the Harbor registry",
				Execute:     inst.docker,
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

// Skill implements the optional command.Skiller interface. Two verbs with no
// arguments render as almost nothing, so the catalog cannot show that both need
// a human: that `auth` only launches a browser and reports success regardless,
// that `docker` logs the whole machine in behind an interactive secret prompt
// and derives its username from GITHUB_TOKEN, or that the prompt's auth
// indicator is inferred from a deliberately failing docker pull.
func (c *Command) Skill(ctx context.Context) string {
	return skill
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) auth(ctx context.Context, r *readline.Readline) error {
	return browser.OpenRawURL(ctx, c.harbor.Config().AuthURL)
}

func (c *Command) docker(ctx context.Context, r *readline.Readline) error {
	client := github.NewClient(
		oauth2.NewClient(
			ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
			),
		),
	)

	var username string

	user, _, err := client.Users.Get(ctx, "")
	if err == nil && user != nil && user.Login != nil {
		username = *user.Login
	} else if username, err = util.Prompt("github username"); err != nil {
		return err
	} else {
		username = strings.TrimSpace(username)
	}

	pterm.Info.Println("registry: " + c.harbor.Config().DockerRegistry())
	pterm.Info.Println("username: " + username)
	pterm.Info.Println("please enter your CLI secret as password provided on your profile settings...")

	_ = browser.OpenRawURL(ctx, c.harbor.Config().AuthURL)

	return shell.New(ctx, c.l, "docker", "login", c.harbor.Config().URL, "-u", username).
		Args(r.AdditionalArgs()...).
		Args(r.AdditionalFlags()...).
		Run()
}
