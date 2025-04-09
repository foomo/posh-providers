package kubeforward

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	gokaziconfig "github.com/foomo/gokazi/pkg/config"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh-providers/kubernets/kubectl"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	ptermx "github.com/foomo/posh/pkg/pterm"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

type (
	Command struct {
		l           log.Logger
		gk          *gokazi.Gokazi
		cfg         Config
		name        string
		configKey   string
		kubectl     *kubectl.Kubectl
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

func NewCommand(l log.Logger, gk *gokazi.Gokazi, kubectl *kubectl.Kubectl, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("kubeforward"),
		gk:        gk,
		name:      "kubeforward",
		configKey: "kubeforward",
		kubectl:   kubectl,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	for key, portForward := range inst.cfg {
		description := fmt.Sprintf("Kubectl port-forward: %s.%s [%s]", portForward.Cluster, portForward.Target, portForward.Port)
		if portForward.Description != "" {
			description = portForward.Description
		}
		inst.gk.Add("kubeforward."+key, gokaziconfig.Task{
			Name:        "kubectl",
			Description: description,
			Path:        "",
			Cwd:         env.ProjectRoot(),
			Args: []string{
				"port-forward", portForward.Target, portForward.Port,
				"--namespace", portForward.Namespace,
			},
		})
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "Manage kubectl port forwards",
		Nodes: []*tree.Node{
			{
				Name:        "connect",
				Description: "Start a port forwaring process",
				Values:      nil,
				Args: tree.Args{
					{
						Name:        "name",
						Description: "Port forward name",
						Repeat:      true,
						Optional:    false,
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.cfg.Names())
						},
					},
				},
				Flags: func(ctx context.Context, r *readline.Readline, fs *readline.FlagSets) error {
					fs.Internal().String("profile", "", "Profile to use")
					fs.Internal().Bool("debug", false, "Debug mode")
					return nil
				},
				Execute: inst.connect,
			},
			{
				Name:        "disconnect",
				Description: "Stop a port forwaring process",
				Values:      nil,
				Args: tree.Args{
					{
						Name:        "name",
						Description: "Port forward name",
						Repeat:      true,
						Optional:    true,
						Suggest: func(ctx context.Context, t tree.Root, r *readline.Readline) []goprompt.Suggest {
							return suggests.List(inst.cfg.Names())
						},
					},
				},
				Execute: inst.disconnect,
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

func (c *Command) disconnect(ctx context.Context, r *readline.Readline) error {
	names := c.cfg.Names()
	if r.Args().LenGt(1) {
		names = r.Args().From(1)
	}
	for _, name := range names {
		pf, ok := c.cfg[name]
		if !ok {
			return errors.Errorf("port forward %s not found", name)
		}

		c.l.Infof("Stopping port forward %s.%s [%s]", pf.Cluster, pf.Target, pf.Port)

		if err := c.gk.Stop(ctx, "kubeforward."+name); errors.Is(err, gokazi.ErrNotRunning) || errors.Is(err, gokazi.ErrNotFound) {
			c.l.Warn("Task: kubeforward." + name + " not running")
		} else if err != nil {
			return errors.Wrap(err, "failed to stop port forward: "+name)
		}
	}

	return nil
}

func (c *Command) connect(ctx context.Context, r *readline.Readline) error {
	fs := r.FlagSets().Default()
	ifs := r.FlagSets().Internal()

	debug, err := ifs.GetBool("debug")
	if err != nil {
		return err
	}

	profile, err := ifs.GetString("profile")
	if err != nil {
		return err
	}

	for _, value := range r.Args().From(1) {
		pf, ok := c.cfg[value]
		if !ok {
			return errors.Errorf("port forward %s not found", value)
		}

		c.l.Infof("Starting port forward %s.%s [%s]", pf.Cluster, pf.Target, pf.Port)

		cmd := exec.CommandContext(ctx, "kubectl",
			"port-forward", pf.Target, pf.Port,
			"--namespace", pf.Namespace,
		)
		cmd.Args = append(cmd.Args, fs.Visited().Args()...)
		cmd.Args = append(cmd.Args, r.AdditionalArgs()...)
		cmd.Env = append(os.Environ(), c.kubectl.Cluster(pf.Cluster).Env(profile))

		if debug {
			cmd.Stderr = ptermx.NewWriter(pterm.Error)
			cmd.Stdout = ptermx.NewWriter(pterm.Info)
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		}

		if err := c.gk.Start(context.WithoutCancel(ctx), "kubeforward."+value, cmd); errors.Is(err, gokazi.ErrAlreadyRunning) {
			c.l.Warn("Task: kubeforward." + value + " already running")
		} else if err != nil {
			return err
		}

		time.Sleep(time.Second)
		if t, err := c.gk.Find(ctx, "kubeforward."+value); err != nil {
			return err
		} else if !t.Running {
			return errors.Errorf("port forward %s not running", value)
		}
	}

	return nil
}
