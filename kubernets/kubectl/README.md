# POSH kubectl provider

## Usage

```go
package main

import (
  "github.com/foomo/posh/provider/foomo/gotsrpc"
  "github.com/foomo/posh/pkg/command"
  "github.com/foomo/posh/pkg/log"
  "github.com/foomo/posh/pkg/plugin"
  "github.com/geschenkidee/opari/cmds/posh/pkg/command/kubectl"
  "github.com/spf13/viper"
)

type Plugin struct {
  l        log.Logger
  kubectl  *kubectl.Kubectl
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    commands: command.Commands{},
  }

  // ...

  inst.commands.Add(gotsrpc.NewCommand(l))

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
