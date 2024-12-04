package migrate

import (
	"context"
	"strconv"

	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/golang-migrate/migrate/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
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
				Name: "database",
				Values: func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.config.Databases())
				},
				Nodes: tree.Nodes{
					{
						Name: "source",
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
								Description: "Roll back the version by 1",
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

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	database := c.config.Database(r.Args().At(0))
	source := c.config.Source(r.Args().At(1))

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
		return m.Up()
	case "up-by-one":
		return m.Steps(1)
	case "down":
		return m.Down()
	case "down-by-one":
		return m.Steps(-1)
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
		return errors.Errorf("unkown command: %s", r.Args().At(2))
	}
}
