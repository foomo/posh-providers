# POSH stern provider

## Usage

### Plugin

```go
package main

import (
  "github.com/foomo/posh/provider/foomo/gotsrpc"
  "github.com/foomo/posh/pkg/command"
  "github.com/foomo/posh/pkg/log"
  "github.com/foomo/posh/pkg/plugin"
  "github.com/spf13/viper"
)

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

  inst.squadron, err = squadron.New(l, inst.kubectl)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create squadron")
  }

  // ...

  inst.commands.Add(stern.NewCommand(l, inst.kubectl, inst.squadron))

  // ...

  return inst, nil
}
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    ## https://github.com/stern/stern/releases
    - name: stern
      tap: foomo/tap/stern/stern
      version: 1.27.0
```
