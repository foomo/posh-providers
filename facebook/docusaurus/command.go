package docusaurus

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/spf13/viper"
)

//go:embed Dockerfile
var dockerfile string

//go:embed SKILL.md
var skill string

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

type (
	// Command implements command.Command, building and running the Docusaurus
	// Docker image for this project's documentation site.
	Command struct {
		l           log.Logger
		cfg         Config
		name        string
		configKey   string
		commandTree tree.Root
	}
	// CommandOption configures a Command built by NewCommand.
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// CommandWithName overrides the command's registered name, which defaults to
// "docusaurus".
func CommandWithName(v string) CommandOption {
	return func(o *Command) {
		o.name = v
	}
}

// CommandWithConfigKey overrides the viper key the Config is unmarshalled
// from, which defaults to "docusaurus".
func CommandWithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// NewCommand builds the docusaurus Command, reading its Config from viper. It
// returns an error if the config cannot be unmarshalled.
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

// Name returns the command's registered name.
func (c *Command) Name() string {
	return c.commandTree.Node().Name
}

// Description returns the command's one-line description.
func (c *Command) Description() string {
	return c.commandTree.Node().Description
}

// Complete returns shell completion suggestions for the current input.
func (c *Command) Complete(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
	return c.commandTree.Complete(ctx, r)
}

// Execute runs the command tree against the parsed arguments and flags.
func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

// Help returns the command's usage text.
func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. This command has no
// args, flags or subcommands, so the renderer would skip it entirely without this
// prose - and what needs saying is that the bare verb blocks on a foreground
// docker run needing a TTY, and overwrites the image tag named in the config.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the documentation-site symptoms that should pull this in, and the fact
// that it serves rather than builds, because the blocking foreground run is the
// thing an agent most needs warning about before it starts one.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when previewing or serving this project's documentation site locally - " +
			"starting the docs dev server, viewing docs in a browser, or working on a Docusaurus " +
			"site. Note it blocks in the foreground on a docker run needing a TTY and overwrites " +
			"the image tag named in config, so it is not usable unattended.",
	}
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
