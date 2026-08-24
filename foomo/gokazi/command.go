package gokazi

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	gokaziconfig "github.com/foomo/gokazi/pkg/config"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh/pkg/command"
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
				Description: "Stop running processes; all of them if no name is given",
				Args: tree.Args{
					{
						Name:        "name",
						Description: "Process name; omit to stop every task in the registry",
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

// Describe implements the optional command.Describer interface, letting
// `posh agent catalog` describe this command's subtree.
func (c *Command) Describe(ctx context.Context) command.CommandInfo {
	return c.commandTree.Describe(ctx)
}

// Skill implements the optional command.Skiller interface. The catalog shows
// `stop` taking an optional repeated name; what it cannot show is that the
// registry is shared with every other provider that starts background tasks, nor
// that omitting the name widens the scope to every task rather than narrowing it.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The triggers are
// deliberately about background processes in general rather than this tool: the
// registry is shared, so this is the command that answers "what is still
// running" for tunnels and forwards other providers started.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when inspecting or stopping the background processes this project " +
			"started - port forwards, SSH or proxy tunnels - or when one is stuck, leaked " +
			"from an earlier shell, or holding a port a new process needs. Covers the " +
			"shared task registry that kubeforward, ssh, dockprox and gost register into.",
	}
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
	if r.Args().LenGt(1) {
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
