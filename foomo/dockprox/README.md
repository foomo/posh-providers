# POSH dockprox provider

Adds a `dockprox` command to manage dockprox processes (start, stop, menubar).

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
  inst.commands.MustAdd(dockprox.NewCommand(l, inst.gk))

  // add checker
  inst.commands.Add(command.NewCheck(l,
    dockprox.Checker(),
  ))

  // ...

  return inst, nil
}
```

The command is configured through functional options:

- `dockprox.CommandWithName("proxy")` — rename the command (default `dockprox`).
- `dockprox.WithConfigKey("proxy")` — read config from a different key (default `dockprox`).

### Config

```yaml
dockprox:
  # Path to the dockprox config
  config: .posh/config/dockprox.yaml
```
