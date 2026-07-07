# POSH ssh provider

Adds an `ssh` command to manage SSH port forwards and SOCKS5 tunnels as
background processes via [gokazi](https://github.com/foomo/gokazi).

## Usage

```
Manage ssh

Usage:
  ssh pfw    start|stop [name...]  Manage port forwards
  ssh socks5 start|stop [name...]  Manage socks5 tunnels
```

Each `name` argument is optional and suggested from your configured
`portForwards` / `socks5Tunnels`; when omitted every configured entry is
started or stopped.

## Plugin

```go
package plugin

type Plugin struct {
  l        log.Logger
  gk       *gokazi.Gokazi
  ssh      *ssh.SSH
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    commands: command.Commands{},
  }

  // ...

  inst.gk = gokazi.New(slog.New(l.SlogHandler()))

  inst.ssh, err = ssh.New(l, inst.gk)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create ssh")
  }

  inst.commands.MustAdd(ssh.NewCommand(l, inst.ssh))

  // ...

  return inst, nil
}
```

The provider and command are configured through functional options:

- `ssh.WithConfigKey("ssh")` — read config from a different key (default `ssh`).
- `ssh.CommandWithName("ssh")` — rename the command (default `ssh`).

### Config

```yaml
ssh:
  portForwards:
    my-forward:
      # Local port to bind (0 = auto-assign)
      port: 15432
      # Target server proxy host
      host: bastion.example.com
      # Target server proxy port
      hostPort: 5432
      # SSH server username
      username: my-user
      # Path to private key (-i)
      identityFile: ~/.ssh/id_ed25519
      # Agent socket path (-o IdentityAgent)
      identityAgent: ""
  socks5Tunnels:
    my-tunnel:
      port: 1080
      host: bastion.example.com
      hostPort: 22
      username: my-user
      identityFile: ~/.ssh/id_ed25519
      identityAgent: ""
```
