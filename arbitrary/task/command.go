package task

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

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

func WithConfigKey(v string) CommandOption {
	return func(o *Command) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("task"),
		name:      "task",
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
		Description: "Run make scripts",
		Args: tree.Args{
			{
				Name: "task",
				Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
					return suggests.List(inst.cfg.Names())
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

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	if err := c.executeTask(ctx, r.Args().At(0)); err != nil {
		return err
	}
	c.l.Success("🔧 | done")
	return nil
}

func (c *Command) executeTask(ctx context.Context, taskID string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	task, ok := c.cfg[taskID]
	if !ok {
		return errors.Errorf("task not found: %s", taskID)
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
		if strings.HasPrefix(cmd, "sudo ") {
			sh = exec.CommandContext(ctx, "sudo", "sh", "-c", strings.TrimPrefix(cmd, "sudo "))
		} else {
			sh = exec.CommandContext(ctx, "sh", "-c", cmd)
		}
		sh.Stdin = os.Stdin
		sh.Stdout = os.Stdout
		sh.Stderr = os.Stderr
		sh.Env = os.Environ()
		c.l.Infof("🔧 | [%d|%d] %s: %s", i+1, len(task.Cmds), taskID, cmd)
		if err := sh.Run(); err != nil {
			return err
		}
	}

	return nil
}
