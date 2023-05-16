# POSH teleport provider

## Usage

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
  teleport *teleport.Teleport
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    cache:    &cache.MemoryCache{},
    commands: command.Commands{},
  }

  // ...

  if value, err := kubectl.New(l, inst.cache); err != nil {
    return nil, err
  } else {
    inst.kubectl = value
  }

  if value, err := teleport.NewTeleport(l, inst.cache); err != nil {
    return nil, err
  } else {
    inst.teleport = value
  }

  // ...

  inst.commands.Add(teleport.NewCommand(l, inst.cache, inst.teleport, inst.kubectl))

  // ...

  return inst, nil
}
```

## Environment:

Add this to your environment:

```yaml
env:
  - name: KUBECONFIG
    value: "${PROJECT_ROOT}/.posh/config/kubeconfig.yaml"
```

## Configuration:

```yaml
kubectl:
  path: devops/config/kubectl
```
