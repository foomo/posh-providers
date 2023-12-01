# POSH doctl provider

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
  doctl    *doctl.Doctl
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

  inst.doctl, err = doctl.New(l, inst.cache)
	if err != nil {
    return nil, errors.Wrap(err, "failed to create doctl")
  }

  // ...

  inst.commands.Add(doctl.NewCommand(l, inst.cache, inst.doctl, inst.kubectl))

  // ...

  return inst, nil
}
```

### Config

```yaml
## doctl
doctl:
  configPath: .posh/config/doctl.yaml
  clusters:
    prod:
      name: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    ## https://github.com/digitalocean/doctl/releases
    - name: doctl
      tap: foomo/tap/digitalocean/doctl
      version: 1.100.0
```
