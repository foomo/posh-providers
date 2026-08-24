package postgres

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/foomo/posh-providers/arbitrary/zip"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/pkg/errors"
)

//go:embed SKILL.md
var skill string

// skillName is the placeholder the embedded SKILL.md uses wherever the command's
// own name appears. Skill substitutes the name the command is registered under,
// which is not necessarily the default: a fragment hardcoding the default tells
// an agent to run a command the project may not have.
const skillName = "{{cmd}}"

// ErrZipRequired is returned instead of dereferencing a nil *zip.Zip: the
// provider accepts CommandWithZip as an option, so a project can reach the
// compression paths without having wired one.
var ErrZipRequired = errors.New("compressing a dump requires the zip provider; pass postgres.CommandWithZip(...) when constructing the command")

type (
	Command struct {
		l           log.Logger
		zip         *zip.Zip
		commandTree tree.Root
	}
	CommandOption func(*Command)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithZip(v *zip.Zip) CommandOption {
	return func(o *Command) {
		o.zip = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, opts ...CommandOption) *Command {
	inst := &Command{
		l: l.Named("postgres"),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	connectionFlags := func(fs *readline.FlagSets) {
		fs.Default().String("port", "", "database server port number")
		fs.Default().String("host", "", "database server host or socket directory")
		fs.Default().String("dbname", "", "connect to database name")
		fs.Default().String("username", "", "connect as specified database user")
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        "postgres",
		Description: "Run psql, or dump and restore databases",
		Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
			connectionFlags(fs)
			return nil
		},
		Execute: inst.execute,
		Nodes: tree.Nodes{
			{
				Name:        "dump",
				Description: "Create a database dump",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Default().Bool("verbose", false, "verbose mode")
					// Options controlling the output content:
					fs.Default().Bool("data-only", false, "dump only the data, not the schema")
					fs.Default().Bool("schema-only", false, "dump only the schema, no data")
					fs.Default().Bool("no-owner", false, "skip restoration of object ownership")
					fs.Default().String("format", "", "output file format ")
					// connection
					connectionFlags(fs)
					fs.Internal().Bool("zip", false, "create a zip file")
					fs.Internal().Bool("dump", false, "use dump format")
					fs.Internal().String("zip-cred", "", "configured zip credential name")

					// The zip provider is optional, so its credential names can only
					// be offered when one was wired.
					if inst.zip != nil {
						if err := fs.Internal().SetValues("zip-cred", inst.zip.Config().CredentialNames()...); err != nil {
							return err
						}
					}

					return nil
				},
				Args: tree.Args{
					{
						Name:        "database",
						Description: "Database name to dump",
					},
					{
						Name:        "dirname",
						Description: "Directory the timestamped dump file is written into",
					},
				},
				Execute: inst.dump,
			},
			{
				Name:        "run-cmd",
				Description: "Run a single SQL command with psql",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					connectionFlags(fs)
					return nil
				},
				Args: tree.Args{
					{
						Name:        "command",
						Description: "SQL passed to psql via --command",
					},
				},
				Execute: inst.runCommand,
			},
			{
				Name:        "run-file",
				Description: "Execute commands from file",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					connectionFlags(fs)
					return nil
				},
				Args: tree.Args{
					{
						Name:        "filename",
						Description: "SQL file passed to psql via --file",
					},
				},
				Execute: inst.runFile,
			},
			{
				Name:        "restore",
				Description: "Restore a database dump",
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Default().Bool("verbose", false, "verbose mode")
					// Options controlling the output content:
					fs.Default().Bool("data-only", false, "restore only the data, no schema")
					fs.Default().Bool("clean", false, "clean (drop) database objects before recreating")
					fs.Default().Bool("create", false, "create the target database")
					fs.Default().Bool("exit-on-error", false, "exit on error, default is to continue")
					fs.Default().Bool("schema-only", false, "restore only the schema, no data")
					fs.Default().Bool("no-owner", false, "skip restoration of object ownership")
					connectionFlags(fs)
					fs.Internal().String("zip-cred", "", "configured zip credential name")

					// The zip provider is optional, so its credential names can only
					// be offered when one was wired.
					if inst.zip != nil {
						if err := fs.Internal().SetValues("zip-cred", inst.zip.Config().CredentialNames()...); err != nil {
							return err
						}
					}

					return nil
				},
				Args: tree.Args{
					{
						Name:        "filename",
						Description: "Path to the dump file",
					},
				},
				Execute: inst.restore,
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
// cannot show that the connection target falls back to the environment, that
// `restore --clean` drops objects first, that the dump filename is generated
// rather than chosen, or that compression needs CommandWithZip to have been wired.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the database tasks and the psql/pg_dump/pg_restore binaries, since a
// request arrives as "back up the database" or as the name of the tool.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when working with a Postgres database from this shell - taking a " +
			"backup or dump, restoring one, running a SQL query or a .sql script, or opening " +
			"an interactive psql session. Also for pg_dump, pg_restore and psql themselves, " +
			"or when a dump needs compressing or password-protecting.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "psql").
		Args(r.Args()...).
		Args(r.Flags()...).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) runFile(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "psql").
		Args(r.Flags()...).
		Args("--file", r.Args().At(1)).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) runCommand(ctx context.Context, r *readline.Readline) error {
	return shell.New(ctx, c.l, "psql").
		Args(r.Flags()...).
		Args("--command", r.Args().At(1)).
		Args(r.AdditionalArgs()...).
		Run()
}

func (c *Command) dump(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()
	ifs := r.FlagSets().Internal()

	database := r.Args().At(1)

	dirname := r.Args().At(2)
	if err := os.MkdirAll(dirname, 0700); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s-%s", dirname, database, time.Now().Format("20060102150405"))
	if log.MustGet(ifs.GetBool("dump"))(c.l) {
		filename += ".dump"

		if err := fs.Set("format", "custom"); err != nil {
			return err
		}
	} else {
		filename += ".sql"
	}

	c.l.Info("Creating database dump: " + filename)

	if out, err := shell.New(ctx, c.l, "pg_dump").
		Args(fs.Visited().Args()...).
		Args(r.AdditionalFlags()...).
		Args(database).
		Args(r.AdditionalArgs()...).
		Args(">", filename).
		Output(); err != nil {
		return errors.Wrap(err, string(out))
	}

	if log.MustGet(ifs.GetBool("zip"))(c.l) {
		if c.zip == nil {
			return ErrZipRequired
		}

		c.l.Info("Compressing database dump...")

		if err := c.zip.Create(ctx, filename); err != nil {
			return err
		}
	}

	if cred := log.MustGet(ifs.GetString("zip-cred"))(c.l); cred != "" {
		if c.zip == nil {
			return ErrZipRequired
		}

		c.l.Info("Securing database dump...")

		if err := c.zip.CreateWithPassword(ctx, filename, cred); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) restore(ctx context.Context, r *readline.Readline) error {
	flags := r.Flags()
	filename := r.Args().At(1)

	c.l.Info("Restoring database dump: " + filename)

	if out, err := shell.New(ctx, c.l, "pg_restore").
		Args(flags...).
		Args(r.AdditionalFlags()...).
		Args(r.AdditionalArgs()...).
		Args(filename).
		Output(); err != nil {
		return errors.Wrap(err, string(out))
	}

	return nil
}
