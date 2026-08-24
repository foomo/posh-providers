package task

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
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
		cfg         Config
		cache       cache.Namespace
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

func WithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, cache cache.Cache, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("task"),
		name:      "task",
		cache:     cache.Get("task"),
		configKey: "task",
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
		Description: "Run a configured task, executing its shell commands",
		Args: tree.Args{
			{
				Name:        "task",
				Description: "Task name from the tasks config or the task directory",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					var ret []prompt.Suggest

					for name, task := range inst.tasks() {
						if !task.Hidden {
							ret = append(ret, goprompt.Suggest{Text: name, Description: task.Description})
						}
					}

					return ret
				},
			},
		},
		Execute: inst.execute,
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

// Skill implements the optional command.Skiller interface. A task name reveals
// nothing about what it runs: the prose names the arbitrary-shell and sudo
// escalation, the recursive deps with no cycle guard, and that a precondition
// which succeeds skips the task rather than enabling it.
func (c *Command) Skill(ctx context.Context, name string) string {
	return strings.ReplaceAll(skill, skillName, name)
}

// SkillMetadata implements the optional command.SkillMetadataer interface,
// supplying the frontmatter of this command's generated skill. The description
// leads with the arbitrary-shell nature of a task, because the reason to load
// this file is to learn what a named task will run before running it.
func (c *Command) SkillMetadata(ctx context.Context, name string) command.SkillMetadata {
	return command.SkillMetadata{
		Description: "Use when running one of this project's named tasks - a setup, bootstrap, " +
			"init, seed, migrate or teardown shortcut defined in the posh config - or when " +
			"asking which tasks exist. Load this before running an unfamiliar task: a task is " +
			"arbitrary shell, may escalate with sudo, and pulls in its dependencies.",
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) tasks() map[string]Task {
	return c.cache.Get("tasks", func() any {
		tasks, err := c.cfg.AllTasks()
		if err != nil {
			c.l.Debug(err)
			return map[string]Task{}
		}

		return tasks
	}).(map[string]Task)
}

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	start := time.Now()

	if err := c.executeTask(ctx, r.Args().At(0)); err != nil {
		return err
	}

	c.l.Success("🔧 | done ⏱︎ " + time.Since(start).Truncate(time.Second).String())

	return nil
}

func (c *Command) executeTask(ctx context.Context, taskID string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	task, ok := c.tasks()[taskID]
	if !ok {
		return errors.Errorf("task not found: %s", taskID)
	}

	for i, cmd := range task.Precondition {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var sh *exec.Cmd
		if after, ok := strings.CutPrefix(cmd, "sudo "); ok {
			sh = exec.CommandContext(ctx, "sudo", "sh", "-c", after)
		} else {
			sh = exec.CommandContext(ctx, "sh", "-c", cmd)
		}

		sh.Stdin = os.Stdin
		sh.Stdout = os.Stdout
		sh.Stderr = os.Stderr

		sh.Env = append(os.Environ(), task.Env...)
		if task.Dir != "" {
			sh.Dir = task.Dir
		}

		sh.Env = append(os.Environ(), task.Env...)
		c.l.Infof("🔧 | {%d|%d} %s: %s", i+1, len(task.Precondition), taskID, cmd)

		if err := sh.Run(); err == nil {
			return nil
		} else {
			c.l.Debug(err.Error())
		}
	}

	if task.Prompt != "" {
		if result, err := pterm.DefaultInteractiveConfirm.WithOnInterruptFunc(func() {
			cancel()
		}).Show(task.Prompt); err != nil {
			return err
		} else if ctx.Err() != nil {
			return ctx.Err()
		} else if !result {
			return nil
		}
	}

	for _, dep := range task.Deps {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := c.executeTask(ctx, dep); err != nil {
			return err
		}
	}

	for i, cmd := range task.Cmds {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var sh *exec.Cmd
		if after, ok := strings.CutPrefix(cmd, "sudo "); ok {
			sh = exec.CommandContext(ctx, "sudo", "sh", "-c", after)
		} else {
			sh = exec.CommandContext(ctx, "sh", "-c", cmd)
		}

		sh.Stdin = os.Stdin
		sh.Stdout = os.Stdout
		sh.Stderr = os.Stderr

		sh.Env = append(os.Environ(), task.Env...)
		if task.Dir != "" {
			sh.Dir = task.Dir
		}

		c.l.Infof("🔧 | [%d|%d] %s: %s", i+1, len(task.Cmds), taskID, cmd)

		if err := sh.Run(); err != nil {
			return err
		}
	}

	return nil
}
