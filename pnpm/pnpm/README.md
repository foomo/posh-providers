# POSH pnpm provider

## Usage

### Plugin

```go
package main

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

  // ...

  inst.commands.Add(pnpm.NewCommand(l, inst.cache))

  // ...

  return inst, nil
}
```
