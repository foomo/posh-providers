package cloudflared

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/foomo/posh/pkg/log"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/viper"
)

type (
	Cloudflared struct {
		l         log.Logger
		cfg       Config
		configKey string
	}
	Option func(*Cloudflared) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfigKey(v string) Option {
	return func(o *Cloudflared) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New command
func New(l log.Logger, opts ...Option) (*Cloudflared, error) {
	inst := &Cloudflared{
		l:         l,
		configKey: "cloudflared",
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(inst); err != nil {
				return nil, err
			}
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (t *Cloudflared) Config() Config {
	return t.cfg
}

func (t *Cloudflared) Disonnect(ctx context.Context, access Access) error {
	ps, err := process.Processes()
	if err != nil {
		return err
	}

	for _, p := range ps {
		if value, _ := p.Name(); value == "cloudflared" {
			if cmdline, _ := p.Cmdline(); strings.Contains(cmdline, "--hostname "+access.Hostname) {
				t.l.Info("closing connection", "hostname", access.Hostname, "pid", p.Pid, "port", access.Port)
				return p.Kill()
			}
		}
	}

	return nil
}

func (t *Cloudflared) Connect(ctx context.Context, access Access) error {
	if t.IsConnected(ctx, access) {
		return errors.Errorf("connection already exists: %s", access.Hostname)
	}

	cmd := exec.CommandContext(ctx, "cloudflared",
		"access", access.Type,
		"--hostname", access.Hostname,
		"--url", fmt.Sprintf("127.0.0.1:%d", access.Port),
	)

	cmd.Env = append(os.Environ(), "HOME="+t.Config().Path)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	if cmd.Process != nil {
		t.l.Info("started access", "hostname", access.Hostname, "port", access.Port, "pid", cmd.Process.Pid)
	}

	return nil
}

func (t *Cloudflared) IsConnected(ctx context.Context, access Access) bool {
	list, err := t.List()
	if err != nil {
		return false
	}

	for _, p := range list {
		if strings.Contains(p.Cmdline, "--hostname "+access.Hostname) {
			return true
		}
	}

	return false
}

func (t *Cloudflared) List() ([]Process, error) {
	ps, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var ret []Process

	for _, p := range ps {
		if value, _ := p.Name(); value == "cloudflared" {
			exe, _ := p.Exe()
			cmdline, _ := p.Cmdline()
			ret = append(ret, Process{
				PID:     fmt.Sprintf("%d", p.Pid),
				Exe:     exe,
				Cmdline: cmdline,
			})
		}
	}

	return ret, nil
}
