# POSH gost provider

Adds a `gost` command to manage [gost](https://github.com/go-gost/gost) (GO Simple Tunnel) processes, starting and stopping tunnels from your configured config files.

## Plugin

```go
package plugin

type Plugin struct {
  l        log.Logger
  gk       *gokazi.Gokazi
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    gk:       gokazi.New(slog.New(l.SlogHandler())),
    commands: command.Commands{},
  }

  // ...

  // add command
  cmd, err := gost.NewCommand(l, inst.gk)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create gost command")
  }
  inst.commands.MustAdd(cmd)

  // ...

  return inst, nil
}
```

The command is configured through functional options:

- `gost.CommandWithName("tunnel")` — rename the command (default `gost`).
- `gost.WithConfigKey("tunnel")` — read config from a different key (default `gost`).

### Config

```yaml
gost:
  # Named gost configs, mapped to their config file paths
  local: .posh/config/gost/local.yaml
  staging: .posh/config/gost/staging.yaml
```
