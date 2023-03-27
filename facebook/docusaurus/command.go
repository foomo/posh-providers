package docusaurus

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/spf13/viper"
)

//go:embed Dockerfile
var dockerfile string

type (
	Command struct {
		l           log.Logger
		cfg         Config
		name        string
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

func NewCommand(l log.Logger, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("docusaurus"),
		name:      "docusaurus",
		configKey: "docusaurus",
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
		Description: "Run docusaurus",
		Execute:     inst.execute,
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

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	if err := shell.New(ctx, c.l,
		"docker", "build",
		"--build-arg", fmt.Sprintf("TAG=%s", c.cfg.NodeTag),
		"--build-arg", fmt.Sprintf("SOURCE=%s", c.cfg.SourcePath),
		"-t", fmt.Sprintf("%s:%s", c.cfg.ImageName, c.cfg.ImageTag),
		"-f", "-",
		".",
	).Stdin(strings.NewReader(dockerfile)).Run(); err != nil {
		return err
	}
	cmd := shell.New(ctx, c.l,
		"docker", "run", "-it", "--rm",
		"-p", fmt.Sprintf("%s:%s", c.cfg.LocalPort, "3000"),
	)
	for _, volume := range c.cfg.Volumes {
		cmd.Args("-v", os.ExpandEnv(volume))
	}
	cmd.Args(fmt.Sprintf("%s:%s", c.cfg.ImageName, c.cfg.ImageTag))
	return cmd.Run()
}
