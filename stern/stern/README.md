# POSH stern provider

## Usage

### Plugin

```go
package main

import (
  // stern takes the squadron.Squadron interface; the implementation lives in
  // the v2 module, and both declare `package squadron`, so alias one of them.
  squadronv2 "github.com/foomo/posh-providers/foomo/squadron/v2"
)

type Plugin struct {
  l        log.Logger
  cache    cache.Cache
  kubectl  *kubectl.Kubectl
  squadron *squadronv2.Squadron
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

  inst.squadron, err = squadronv2.New(l, inst.kubectl)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create squadron")
  }

  // ...

  inst.commands.MustAdd(stern.NewCommand(l, inst.kubectl, inst.squadron))

  // ...

  return inst, nil
}
```

### Config

```yaml
stern:
  queries:
    all:
      query: ['.*', '--all-namespaces']
      queries:
        panic:
          query: ['--include', 'panic']
        fatal:
          query: ['--include', '"\"level\":\"fatal\""']
        errors:
          query: ['--include', '"\"level\":\"error\""']
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
