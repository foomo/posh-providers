package sshtunnel

import (
	"context"
	"fmt"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
	"github.com/spf13/viper"
)

// SSHTunnel command
type (
	SSHTunnel struct {
		l         log.Logger
		cfg       Config
		configKey string
	}
	Option func(*SSHTunnel) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithConfigKey(v string) Option {
	return func(o *SSHTunnel) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New command
func New(l log.Logger, opts ...Option) (*SSHTunnel, error) {
	inst := &SSHTunnel{
		l:         l,
		configKey: "sshTunnel",
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

func (s *SSHTunnel) Config() Config {
	return s.cfg
}

func (s *SSHTunnel) SocketsDir() string {
	return s.cfg.SocketsDir
}

func (s *SSHTunnel) TempDir() string {
	return s.cfg.TempDir
}

func (s *SSHTunnel) Tunnel(name string) (Tunnel, bool) {
	return s.cfg.Tunnel(name)
}

// IsTunnelRunning checks if the SSH tunnel with the given name is currently active.
func (s *SSHTunnel) IsTunnelRunning(ctx context.Context, name string) bool {
	tunnel, _ := s.Tunnel(name)
	sudo := ""
	if tunnel.Sudo {
		sudo = "sudo"
	}

	cmdStr := fmt.Sprintf(
		"%s ssh",
		sudo,
	)

	sh := shell.New(ctx, s.l, cmdStr).
		Args("-S", fmt.Sprintf("%s/%s", s.SocketsDir(), tunnel.Name)).
		Args("-O", "check").
		Args(fmt.Sprintf("%s@%s", tunnel.TargetUsername, tunnel.TargetHost)).
		Quiet()

	if err := sh.Run(); err != nil {
		return false
	}

	return true
}

// IsTargetProxyPortOpen checks if the targetHost:targetProxyPort is open.
func (s *SSHTunnel) IsTargetProxyPortOpen(ctx context.Context, name, targetAuthPassword, targetAuthPrivateKey string) bool {
	tunnel, _ := s.Tunnel(name)

	sshpass := ""
	if targetAuthPassword != "" {
		sshpass = fmt.Sprintf("sshpass -p %s", targetAuthPassword)
	}

	cmdStr := fmt.Sprintf("%s ssh", sshpass)
	ncCmdStr := fmt.Sprintf("nc -z -w2 %s %d", tunnel.TargetProxyHost, tunnel.TargetProxyPort)

	// Build the shell command
	sh := shell.New(ctx, s.l, cmdStr).
		Args(fmt.Sprintf("%s@%s", tunnel.TargetUsername, tunnel.TargetHost))

	if targetAuthPrivateKey != "" {
		sh.Args("-i", targetAuthPrivateKey)
	}

	sh.Args(ncCmdStr).
		Quiet()

	return sh.Run() == nil
}

// IsLocalPortInUse checks if the LocalPort is in already use.
func (s *SSHTunnel) IsLocalPortInUse(ctx context.Context, name string) bool {
	tunnel, _ := s.Tunnel(name)
	cmdStr := fmt.Sprintf("nc -z -w2 127.0.0.1 %d", tunnel.LocalPort)

	sh := shell.New(ctx, s.l, cmdStr)
	sh.Quiet()
	err := sh.Run()
	return err == nil
}
