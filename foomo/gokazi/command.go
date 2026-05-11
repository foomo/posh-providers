package gokazi

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"

	gokaziconfig "github.com/foomo/gokazi/pkg/config"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type (
	Command struct {
		l           log.Logger
		gk          *gokazi.Gokazi
		cfg         Config
		configKey   string
		name        string
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

func NewCommand(l log.Logger, gk *gokazi.Gokazi, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("gokazi"),
		gk:        gk,
		configKey: "gokazi",
		name:      "gokazi",
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
		Description: "Manage gokazi processes",
		Nodes: tree.Nodes{
			{
				Name:        "list",
				Description: "List processes",
				Execute:     inst.list,
			},
			{
				Name:        "stop",
				Description: "Stop process",
				Args: tree.Args{
					{
						Name:        "name",
						Description: "Process name",
						Repeat:      true,
						Optional:    true,
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							tasks, err := inst.gk.List(ctx)
							if err != nil {
								return nil
							}

							return suggests.List(slices.Sorted(maps.Keys(tasks)))
						},
					},
				},
				Execute: inst.stop,
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

func (c *Command) Shutdown(ctx context.Context) error {
	if !c.cfg.Cleanup {
		return nil
	}

	t, err := c.gk.List(ctx)
	if err != nil {
		return err
	}

	c.gk.Add("posh", gokaziconfig.Task{
		Name:        "posh",
		Description: "",
		Path:        "",
		Cwd:         env.ProjectRoot(),
	})

	_, err = c.gk.Find(ctx, "posh")
	if errors.Is(err, gokazi.ErrMultipleFound) {
		return nil
	} else if err != nil {
		return err
	}

	wg, _ := errgroup.WithContext(ctx)

	c.l.Debug("Shutting down task")

	for name, task := range t {
		if task.Pid > 0 {
			wg.Go(func() error {
				c.l.Debugf("└  "+name+" (%d)", task.Pid)

				process, err := os.FindProcess(int(task.Pid))
				if err != nil {
					return err
				}

				defer func() {
					if err := process.Release(); err != nil {
						c.l.Warn("failed to release process", err.Error())
					}
				}()

				return process.Signal(os.Interrupt)
			})
		}
	}

	return wg.Wait()
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) list(ctx context.Context, r *readline.Readline) error {
	tasks, err := c.gk.List(ctx)
	if err != nil {
		return err
	}

	t := pterm.DefaultTable.WithWriter(os.Stdout).WithHasHeader(true)
	t.Data = append(t.Data, []string{"PID", "NAME", "RUNNING"})

	names := slices.Sorted(maps.Keys(tasks))
	for _, name := range names {
		task := tasks[name]
		t.Data = append(t.Data, []string{fmt.Sprint(task.Pid), name, fmt.Sprint(task.Running)})
	}

	return t.Render()
}

func (c *Command) stop(ctx context.Context, r *readline.Readline) error {
	tasks, err := c.gk.List(ctx)
	if err != nil {
		return err
	}

	names := slices.Sorted(maps.Keys(tasks))
	if r.Args().LenGt(2) {
		names = r.Args().From(1)
	}

	for _, name := range names {
		if !tasks[name].Running {
			c.l.Info("Skip:", name)
			continue
		}

		if err := c.gk.Stop(ctx, name); err != nil {
			c.l.Error("Failed:", name, "pid", tasks[name].Pid, "err", err.Error())
			continue
		}

		c.l.Success("Stopped:", name, "pid", tasks[name].Pid)
	}

	return nil
}
