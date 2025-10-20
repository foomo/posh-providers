package harbor

import (
	"context"
	"os"

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
		Description: "Run harbor",
		Execute:     inst.auth,
		Nodes: tree.Nodes{
			{
				Name:        "auth",
				Args:        nil,
				Description: "Sign in to Harbor",
				Execute:     inst.auth,
			},
			{
				Name:        "docker",
				Args:        nil,
				Description: "Configure docker to be able to access registry.",
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
