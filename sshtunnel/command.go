package sshtunnel

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command/tree"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/goprompt"
	"github.com/foomo/posh/pkg/readline"
	"github.com/foomo/posh/pkg/shell"
	"github.com/foomo/posh/pkg/util/suggests"
	"github.com/pkg/errors"
)

const (
	AuthTypePass = "sshpass"
	AuthTypeKey  = "key"
)

type (
	Command struct {
		l           log.Logger
		cache       cache.Namespace
		name        string
		commandTree tree.Root
		op          *onepassword.OnePassword
		sshTunnel   *SSHTunnel
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

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewCommand(l log.Logger, sshTunnel *SSHTunnel, cache cache.Cache, op *onepassword.OnePassword, opts ...CommandOption) (*Command, error) {
	inst := &Command{
		l:         l.Named("SSHTunnel"),
		name:      "sshtunnel",
		op:        op,
		sshTunnel: sshTunnel,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	inst.l = l.Named(inst.name)
	inst.cache = cache.Get(inst.name)

	tunnelsValues := func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
		var ret []string
		ret = append(ret, inst.sshTunnel.cfg.TunnelNames()...)
		return suggests.List(ret)
	}

	cmdsValues := func(ctx context.Context, r *readline.Readline) []goprompt.Suggest {
		return []goprompt.Suggest{
			{Text: "start", Description: "Start SSH-Tunnel"},
			{Text: "stop", Description: "Stop SSH-Tunnel"},
		}
	}

	inst.commandTree = tree.New(&tree.Node{
		Name:        inst.name,
		Description: "SSH-Tunnels Managament",
		Nodes: tree.Nodes{
			{
				Name:        "tunnel",
				Description: "Tunnel name",
				Values:      tunnelsValues,
				Nodes: tree.Nodes{
					{
						Name:        "cmd",
						Description: "Command type",
						Values:      cmdsValues,
						Execute:     inst.execute,
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

func (c *Command) Validate(ctx context.Context, r *readline.Readline) error {
	switch {
	case r.Args().LenIs(0):
		return errors.New("missing [tunnel] argument")
	case r.Args().LenIs(1):
		return errors.New("missing [cmd] argument")
	}
	return nil
}

func (c *Command) Execute(ctx context.Context, r *readline.Readline) error {
	return c.commandTree.Execute(ctx, r)
}

func (c *Command) Help(ctx context.Context, r *readline.Readline) string {
	return c.commandTree.Help(ctx, r)
}

// ------------------------------------------------------------------------------------------------
// ~ Execution
// ------------------------------------------------------------------------------------------------
func (c *Command) execute(ctx context.Context, r *readline.Readline) error {
	tunnelName, cmd := r.Args()[0], r.Args()[1]

	tunnelConfig, ok := c.sshTunnel.Tunnel(tunnelName)
	if !ok {
		return errors.Errorf("tunnel configuration %q not found", tunnelName)
	}

	tempDir := c.sshTunnel.TempDir()
	socketsDir := c.sshTunnel.SocketsDir()
	if err := os.MkdirAll(socketsDir, 0o755); err != nil {
		return errors.Wrap(err, "failed to create SSH sockets directory")
	}
	socketsFile := fmt.Sprintf("%s/%s", socketsDir, tunnelName)

	sudo := ""
	if tunnelConfig.Sudo {
		sudo = "sudo"
	}

	var cmdStr string
	var sshpass string

	switch cmd {
	case "start":
		if c.sshTunnel.IsTunnelRunning(ctx, tunnelName) {
			c.l.Infof("Tunnel %s is already running", tunnelName)
			return nil
		}

		if c.sshTunnel.IsLocalPortInUse(ctx, tunnelName) {
			return errors.Errorf("bind [127.0.0.1]:%d: Address already in use", tunnelConfig.LocalPort)
		}

		var (
			targetAuthPassword   string
			targetAuthPrivateKey string
			cleanup              func()
			err                  error
		)
		if tunnelConfig.TargetAuth.Type == AuthTypePass {
			targetAuthPassword, _, err = c.resolveSSHCredential(ctx, tunnelConfig.TargetAuth.Password, tempDir, AuthTypePass)
			if err != nil {
				return err
			}
		}

		if tunnelConfig.TargetAuth.Type == AuthTypeKey {
			targetAuthPrivateKey, cleanup, err = c.resolveSSHCredential(ctx, tunnelConfig.TargetAuth.PrivateKey, tempDir, AuthTypeKey)
			if cleanup != nil {
				defer cleanup()
			}
			if err != nil {
				return err
			}
		}

		if !c.sshTunnel.IsTargetProxyPortOpen(ctx, tunnelName, targetAuthPassword, targetAuthPrivateKey) {
			return errors.Errorf(
				"cannot reach target proxy port %d at %s from host %s",
				tunnelConfig.TargetProxyPort,
				tunnelConfig.TargetProxyHost,
				tunnelConfig.TargetHost,
			)
		}

		if targetAuthPassword != "" {
			sshpass = fmt.Sprintf("sshpass -p %s", targetAuthPassword)
		}
		cmdStr = fmt.Sprintf(
			"%s %s ssh",
			sudo,
			sshpass,
		)

		sh := shell.New(ctx, c.l, cmdStr).
			Args("-M").
			Args("-S", socketsFile).
			Args("-fnNT").
			Args("-L", fmt.Sprintf(
				"%d:%s:%d",
				tunnelConfig.LocalPort,
				tunnelConfig.TargetProxyHost,
				tunnelConfig.TargetProxyPort,
			)).
			Args(fmt.Sprintf("%s@%s", tunnelConfig.TargetUsername, tunnelConfig.TargetHost))

		if targetAuthPrivateKey != "" {
			sh.Args("-i", targetAuthPrivateKey)
		}

		if err := sh.Run(); err != nil {
			return errors.Wrap(err, "failed to start ssh tunnel")
		}

		c.l.Successf("Tunnel %s started successfully", tunnelName)

	case "stop":
		if !c.sshTunnel.IsTunnelRunning(ctx, tunnelName) {
			c.l.Infof("Tunnel %s is not running", tunnelName)
			return nil
		}

		cmdStr = fmt.Sprintf(
			"%s ssh",
			sudo,
		)

		sh := shell.New(ctx, c.l, cmdStr).
			Args("-S", socketsFile).
			Args("-O", "exit").
			Args(fmt.Sprintf("%s@%s", tunnelConfig.TargetUsername, tunnelConfig.TargetHost))

		if err := sh.Run(); err != nil {
			return errors.Wrap(err, "failed to stop ssh tunnel")
		}

		c.l.Successf("Tunnel %s stopped successfully", tunnelName)

	default:
		return errors.Errorf("unknown command: %s", cmd)
	}

	return nil
}

// resolveSSHCredential resolves SSH credentials for a tunnel.
// - For "sshpass": returns the password.
// - For "key": returns a file path to an existing key or writes content to a temp file.
// Supports resolving secrets from 1Password.
func (c *Command) resolveSSHCredential(ctx context.Context, value, tempDir, authType string) (string, func(), error) {
	value = strings.TrimSpace(value)
	if value == "" {
		switch authType {
		case AuthTypePass:
			return "", nil, errors.New("password field is required for authType sshpass")
		case AuthTypeKey:
			return "", nil, errors.New("privateKey field is required for authType key")
		}
		return "", nil, nil
	}
	isFrom1Password := false

	if strings.HasPrefix(value, "<% op") {
		isFrom1Password = true
		if c.op == nil {
			return "", nil, errors.New("1Password client not available")
		}
		if ok, _ := c.op.IsAuthenticated(ctx); !ok {
			c.l.Info("Missing 1Password session, please login")
			if err := c.op.SignIn(ctx); err != nil {
				return "", nil, errors.Wrap(err, "failed to sign in to 1Password")
			}
		}
		secretBytes, err := c.op.Render(ctx, value)
		if err != nil {
			return "", nil, errors.Wrap(err, "failed to render 1Password secret")
		}
		value = strings.TrimSpace(string(secretBytes))
	}

	switch authType {
	case "sshpass":
		return value, nil, nil

	case "key":
		// Check if the value is a path to an existing file
		if !isFrom1Password {
			if _, err := os.Stat(value); err == nil {
				return value, nil, nil
			} else if os.IsNotExist(err) {
				return "", nil, errors.Wrap(err, "private key file does not exist")
			}
		}
		// otherwise treat value as key content and write it to a temp file
		tmpFile, err := os.CreateTemp(tempDir, "sshkey-*")
		if err != nil {
			return "", nil, errors.Wrap(err, "failed to create temporary key file")
		}
		defer tmpFile.Close()

		if _, err := tmpFile.WriteString(value); err != nil {
			return "", nil, errors.Wrap(err, "failed to write private key to temporary file")
		}

		cleanup := func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				c.l.Warnf("failed to remove temp file %s: %v", tmpFile.Name(), err)
			}
		}
		return tmpFile.Name(), cleanup, nil

	default:
		return "", nil, errors.Errorf("unknown tunnel auth type: %s", authType)
	}
}
