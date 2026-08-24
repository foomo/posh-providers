package zip

import (
	"context"
	_ "embed"
	"strings"

	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/files"
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
		zip         *Zip
		commandTree tree.Root
	}
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, zip *Zip, opts ...CommandOption) *Command {
	inst := &Command{
		l:   l.Named("zip"),
		zip: zip,
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "zip",
		Description: "Extract zip archives, optionally with a 1Password-held password",
		Nodes: tree.Nodes{
			{
				Name:        "extract",
				Description: "Extract a zip archive into the directory containing it",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("cred", "", "Credential name from the zip config, whose 1Password password is used")

					if err := fs.Internal().SetValues("cred", inst.zip.Config().CredentialNames()...); err != nil {
						return err
					}

					return nil
				},
				Args: tree.Args{
					{
						Name:        "filename",
						Description: "Path to the zip file; it is extracted into its own directory",
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							ret, _ := files.Find(ctx, ".", "*.zip",
								files.FindWithIgnore(`^\.`, "vendor", "node_modules"),
							)

							return suggests.List(ret)
						},
					},
				},
				Execute: inst.extract,
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

// Skill implements the optional command.Skiller interface. The rendered tree
// cannot show that extraction happens in the archive's own directory and can
// overwrite files there, that `--cred` resolves a 1Password secret and hands it
// to unzip on the command line, or that the provider's Create methods are not
// reachable from the shell at all.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the password-protected case explicitly, since a plain unzip needs no
// posh command and the 1Password-held archive password is the reason to be here.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when unpacking a .zip archive in this project, especially a " +
			"password-protected one whose password lives in 1Password - encrypted database " +
			"dumps, backups or exports. Also when an unzip prompts for a password, or when " +
			"asking which archive credentials the project has configured.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) extract(ctx context.Context, r *readline.Readline) error {
	ifs := r.FlagSets().Internal()
	filename := r.Args().At(1)

	if cred := log.MustGet(ifs.GetString("cred"))(c.l); cred != "" {
		return c.zip.ExtractWithPassword(ctx, filename, cred)
	}

	return c.zip.Extract(ctx, filename)
}
