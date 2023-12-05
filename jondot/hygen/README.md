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

  inst.commands.MustAdd(hygen.NewCommand(l, inst.cache))

  // ...

  return inst, nil
}
```

### Config

```yaml
## hygen
hygen:
  templatePath: .posh/scaffold
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    ## https://github.com/jondot/hygen/releases
    - name: hygen
      tap: foomo/tap/jondot/hygen
      version: 6.2.11
```
