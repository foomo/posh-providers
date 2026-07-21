# POSH ku provider

> A fast, keyboard-driven Kubernetes TUI — see [github.com/bjarneo/ku](https://github.com/bjarneo/ku).

## Usage

### Plugin

```go
package main

type Plugin struct {
  l        log.Logger
  cache    cache.Cache
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

  inst.kubectl, err = kubectl.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create kubectl")
  }

  // ...

  inst.commands.MustAdd(ku.NewCommand(l, inst.kubectl))

  // ... or wire squadron for fleet/namespace resolution:
  // inst.commands.MustAdd(ku.NewCommand(l, inst.kubectl, ku.CommandWithSquadron(inst.squadron)))

  // ...

  return inst, nil
}
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    ## https://github.com/bjarneo/ku/releases
    - name: ku
      tap: foomo/tap/bjarneo/ku
      version: 0.1.0
```
