# POSH stackit provider

## Usage

### Plugin

```go
package main

import (
  "github.com/foomo/posh-providers/kubernetes/kubectl"
  "github.com/foomo/posh-providers/stackitcloud/stackit"
  "github.com/foomo/posh/pkg/cache"
  "github.com/foomo/posh/pkg/command"
)

type Plugin struct {
  l        log.Logger
  cache    cache.Cache
  stackit  *stackit.Stackit
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

  inst.stackit, err = stackit.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create stackit")
  }

  // ...

  inst.commands.Add(stackit.NewCommand(l, inst.cache, inst.stackit, inst.kubectl))

  // ...

  return inst, nil
}

```

### Config

```yaml
## stackit
stackit:
  projects:
    my-project:
      id: 123456-123456-123456
      clusters:
        dev:
          name: my-project-dev-cluster
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    ## https://github.com/stackitcloud/stackit-cli/releases
    - name: stackit
      tap: foomo/tap/stackitcloud/stackit-cli
      version: 0.9.0
```
