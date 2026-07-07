# POSH kubeforward provider

Adds a `kubeforward` command to manage named `kubectl port-forward` processes,
starting and stopping them as background tasks via gokazi.

## Plugin

```go
package plugin

type Plugin struct {
  l        log.Logger
  cache    cache.Cache
  gokazi   *gokazi.Gokazi
  kubectl  *kubectl.Kubectl
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  var err error
  inst := &Plugin{
    l:        l,
    cache:    &cache.MemoryCache{},
    commands: command.Commands{},
  }

  // ...

  inst.gokazi = gokazi.New(slog.New(l.SlogHandler()))

  inst.kubectl, err = kubectl.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create kubectl")
  }

  // ...

  cmd, err := kubeforward.NewCommand(l, inst.gokazi, inst.kubectl)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create kubeforward command")
  }
  inst.commands.MustAdd(cmd)

  // ...

  return inst, nil
}
```

The command is configured through functional options:

- `kubeforward.CommandWithName("forward")` — rename the command (default `kubeforward`).
- `kubeforward.CommandWithConfigKey("forward")` — read config from a different key (default `kubeforward`).

### Config

```yaml
kubeforward:
  # Named port forward, suggested as an argument to connect/disconnect
  my-database:
    # Target cluster
    cluster: my-cluster
    # Target namespace
    namespace: my-namespace
    # Optional description (defaults to a generated one)
    description: Forward to my database
    # Target name (e.g. service/foo, pod/bar)
    target: service/my-database
    # Target and host port mapping
    port: "5432:5432"
```
