package migrate

import (
	"context"
	_ "embed"
	"strconv"
	"strings"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/golang-migrate/migrate/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
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
		op          *onepassword.OnePassword
		name        string
		config      Config
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

func CommandWithOnePassword(v *onepassword.OnePassword) CommandOption {
	return func(o *Command) {
		o.op = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("migrate"),
		name:      "migrate",
		configKey: "migrate",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.config); err != nil {
		return nil, err
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Manage database migrations",
		Nodes: tree.Nodes{
			{
				Name:        "database",
				Description: "Name of a database from the `migrate.databases` config key",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.config.Databases())
				},
				Nodes: tree.Nodes{
					{
						Name:        "source",
						Description: "Name of a migration source from the `migrate.sources` config key",
						Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.config.Sources())
						},
						Nodes: tree.Nodes{
							{
								Name:        "up",
								Description: "Migrate the DB to the most recent version available",
								Execute:     inst.execute,
							},
							{
								Name:        "up-by-one",
								Description: "Migrate the DB up by 1",
								Execute:     inst.execute,
							},
							{
								Name:        "down",
								Description: "Roll back all migrations",
								Execute:     inst.execute,
							},
							{
								Name:        "down-by-one",
								Description: "Migrate the DB down by 1",
								Execute:     inst.execute,
							},
							{
								Name:        "force",
								Description: "Sets a migration version",
								Args: tree.Args{
									{
										Name:        "version",
										Description: "Version to migrate",
									},
								},
								Execute: inst.execute,
							},
							{
								Name:        "migrate",
								Description: "Migrates either up or down to the specified version",
								Args: tree.Args{
									{
										Name:        "version",
										Description: "Version to migrate",
									},
								},
								Execute: inst.execute,
							},
							{
								Name:        "version",
								Description: "Print the current version of the database",
								Execute:     inst.execute,
							},
							{
								Name:        "drop",
								Description: "Deletes everything in the database",
								Execute:     inst.execute,
							},
						},
					},
				},
			},
		},
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog names
// what each verb does but cannot show that `drop` and `down` are irreversible
// and unconfirmed, that the database they hit is picked by the first argument,
// or that the two placeholders resolve against config keys rather than the
// database itself.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// names the symptoms - a dirty version, a schema behind the migration files -
// because that is how a migration problem is usually reported.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when changing or inspecting a database schema through migrations - " +
			"applying pending migrations, checking the current migration version, rolling back, " +
			"jumping to a specific version, recovering a database stuck in a dirty state, or " +
			"dropping a database. Also when a golang-migrate migration errors.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	database := c.config.Database(r.Args().At(0))
	source := c.config.Source(r.Args().At(1))

	if c.op != nil {
		if out, err := c.op.Render(ctx, database); err != nil {
			return err
		} else {
			database = string(out)
		}
	}

	m, err := migrate.New(source, database)
	if err != nil {
		return err
	}

	m.Log = &logger{l: c.l}

	go func() {
		if <-ctx.Done(); true {
			c.l.Info("triggering graceful migration shutdown")

			m.GracefulStop <- true
		}
	}()

	defer func() {
		if dErr, sErr := m.Close(); dErr != nil {
			c.l.Warn(dErr)
		} else if sErr != nil {
			c.l.Warn(sErr)
		}
	}()

	switch r.Args().At(2) {
	case "up":
		err := m.Up()
		if errors.Is(err, migrate.ErrNoChange) {
			c.l.Info("no change")
			return nil
		}

		return err
	case "up-by-one":
		err := m.Steps(1)
		if errors.Is(err, migrate.ErrNoChange) {
			c.l.Info("no change")
			return nil
		}

		return err
	case "down":
		return m.Down()
	case "down-by-one":
		return m.Steps(-1)
	case "migrate":
		i, err := strconv.Atoi(r.Args().At(3))
		if err != nil {
			return err
		}

		return m.Migrate(uint(i))
	case "force":
		i, err := strconv.Atoi(r.Args().At(3))
		if err != nil {
			return err
		}

		return m.Force(i)
	case "drop":
		return m.Drop()
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			return err
		}

		c.l.Infof("Version: %d, Dirty: %t", version, dirty)

		return nil
	default:
		return errors.Errorf("unknown command: %s", r.Args().At(2))
	}
}
