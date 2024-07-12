package zip

import (
	"context"

	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/foomo/posh/pkg/util/suggests"
)

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
		Description: "Zip command",
		Nodes: tree.Nodes{
			{
				Name: "extract",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("cred", "", "configured zip credential name")
					if err := fs.Internal().SetValues("cred", inst.zip.Config().CredentialNames()...); err != nil {
						return err
					}
					return nil
				},
				Args: tree.Args{
					{
						Name:        "filename",
						Description: "Path to zip file",
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
