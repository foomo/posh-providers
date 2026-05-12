package ssh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	gonet "github.com/foomo/go/net"
	goos "github.com/foomo/go/os"
	gokaziconfig "github.com/foomo/gokazi/pkg/config"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

// PortForward command
type (
	SSH struct {
		l         log.Logger
		gk        *gokazi.Gokazi
		cfg       Config
		configKey string
	}
	Option func(*SSH) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfigKey(v string) Option {
	return func(o *SSH) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New command
func New(l log.Logger, gk *gokazi.Gokazi, opts ...Option) (*SSH, error) {
	inst := &SSH{
		l:         l,
		gk:        gk,
		configKey: "ssh",
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

	for name, value := range inst.cfg.PortForwards {
		inst.gk.Add("ssh.pfw."+name, gokaziconfig.Task{
			Name:        "ssh",
			Description: fmt.Sprintf("fwd: %d:%s:%d", value.Port, value.Host, value.HostPort),
			Path:        "",
			Cwd:         env.ProjectRoot(),
			Args: []string{
				"-L", fmt.Sprintf("%d:%s:%d", value.Port, value.Host, value.HostPort),
			},
		})
	}

	for name, value := range inst.cfg.Socks5Tunnels {
		inst.gk.Add("ssh.socks5."+name, gokaziconfig.Task{
			Name:        "ssh",
			Description: fmt.Sprintf("%d:%s:%d", value.Port, value.Host, value.HostPort),
			Path:        "",
			Cwd:         env.ProjectRoot(),
			Args: []string{
				"-D", value.Host,
			},
		})
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (s *SSH) Config() Config {
	return s.cfg
}

func (s *SSH) StartPortForward(ctx context.Context, name string) error {
	c, ok := s.cfg.PortForward(name)
	if !ok {
		return fmt.Errorf("config %s not found", name)
	}

	port := c.Port
	if port <= 0 {
		if value, err := gonet.FreePort(ctx); err != nil {
			return err
		} else {
			port = value
		}
	}

	cmd := exec.CommandContext(ctx, "ssh",
		"-L", fmt.Sprintf("%d:%s:%d", c.Port, c.Host, c.HostPort),
		// "-f", // Run in background after authentication
		"-M", // Places the ssh client into “master” mode for connection sharing
		"-n", // Redirects stdin from /dev/null (actually, prevents reading from stdin)
		"-N", // Do not execute remote command (for forwarding only)
		"-T", // Disable pseudo-terminal allocation
		// "-v", // Verbose mode
	)
	cmd.Env = os.Environ()
	// cmd.Stdout = pterm.Debug.Writer
	// cmd.Stderr = pterm.Error.Writer
	// cmd.WaitDelay = 3 * time.Second

	if c.IdentityFile != "" {
		value, err := goos.Expand(c.IdentityFile)
		if err != nil {
			return err
		}

		cmd.Args = append(cmd.Args, "-i", value)
	}

	if c.IdentityAgent != "" {
		cmd.Args = append(cmd.Args, "-o", "IdentityAgent="+c.IdentityAgent)
	}

	if c.Username != "" {
		cmd.Args = append(cmd.Args, "-l", c.Username)
	}

	cmd.Args = append(cmd.Args, os.ExpandEnv(c.Host))

	err := s.gk.Start(context.WithoutCancel(ctx), "ssh.pfw."+name, cmd)
	if errors.Is(err, gokazi.ErrAlreadyRunning) {
		return nil
	} else if err != nil {
		return err
	}

	pid := cmd.Process.Pid
	addr := fmt.Sprintf("localhost:%d", port)

	pterm.Info.Println("SSH Port Forward ready at", addr, "pid", pid)

	return nil
}

func (s *SSH) StopPortForward(ctx context.Context, name string) error {
	err := s.gk.Stop(context.WithoutCancel(ctx), "ssh.pfw."+name)
	if errors.Is(err, gokazi.ErrNotRunning) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func (s *SSH) StartSocks5Tunnel(ctx context.Context, name string) error {
	c, ok := s.cfg.Socks5Tunnel(name)
	if !ok {
		return fmt.Errorf("SOCKS proxy %s not found", name)
	}

	port := c.Port
	if port <= 0 {
		if value, err := gonet.FreePort(ctx); err != nil {
			return err
		} else {
			port = value
		}
	}

	cmd := exec.CommandContext(ctx, "ssh",
		"-D", fmt.Sprintf("%d", port),
		"-N", // Do not execute remote command (for forwarding only)
		// "-f", // Run in background after authentication
		"-o", "BatchMode=yes",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=10",
		"-o", "ServerAliveCountMax=3",
		// "-v", // Verbose mode.
	)
	cmd.Env = os.Environ()

	if c.HostPort > 0 {
		cmd.Args = append(cmd.Args, "-p", fmt.Sprintf("%d", c.HostPort))
	}

	if c.IdentityFile != "" {
		value, err := goos.Expand(c.IdentityFile)
		if err != nil {
			return err
		}

		cmd.Args = append(cmd.Args, "-i", value)
	}

	if c.IdentityAgent != "" {
		cmd.Args = append(cmd.Args, "-o", "IdentityAgent="+c.IdentityAgent)
	}

	if c.Username != "" {
		cmd.Args = append(cmd.Args, "-l", os.ExpandEnv(c.Username))
	}

	cmd.Args = append(cmd.Args, os.ExpandEnv(c.Host))

	err := s.gk.Start(context.WithoutCancel(ctx), "ssh.socks5."+name, cmd)
	if errors.Is(err, gokazi.ErrAlreadyRunning) {
		return nil
	} else if err != nil {
		return err
	}

	pid := cmd.Process.Pid
	addr := fmt.Sprintf("localhost:%d", port)

	pterm.Success.Println("SSH socks5 tunnel ready at", addr, "pid", pid)

	return nil
}

func (s *SSH) StopSocks5Tunnel(ctx context.Context, name string) error {
	err := s.gk.Stop(context.WithoutCancel(ctx), "ssh.socks5."+name)
	if errors.Is(err, gokazi.ErrNotRunning) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}
