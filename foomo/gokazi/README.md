# POSH gokazi provider

Adds a `gokazi` command to list and stop background processes managed by
[gokazi](https://github.com/foomo/gokazi). The shared `*gokazi.Gokazi` instance
is also consumed by other providers (`ssh`, `dockprox`, `gost`, `kubeforward`)
to run their background tasks.

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
    commands: command.Commands{},
  }

  // ...

  inst.gk = gokazi.New(slog.New(l.SlogHandler()))

  // add checker
  inst.commands.Add(command.NewCheck(l, gokazi.TasksChecker(inst.gk)))

  // add command
  cmd, err := gokazi.NewCommand(l, inst.gk)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create gokazi command")
  }
  inst.commands.Add(cmd)

  // ...

  return inst, nil
}
```

The command is configured through functional options:

- `gokazi.CommandWithName("gokazi")` — rename the command (default `gokazi`).
- `gokazi.CommandWithConfigKey("gokazi")` — read config from a different key (default `gokazi`).

### Config

```yaml
gokazi:
  # Cleanup will stop all processes if last posh is closed
  cleanup: false
```
