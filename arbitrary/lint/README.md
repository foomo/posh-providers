# POSH lint provider

> Adds configured linters to your project.

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

  // Linters are discovered from the command registry by type assertion: any
  // registered command implementing Lint(ctx, fix) is picked up automatically,
  // so add this last, after the linting providers themselves.
  inst.commands.Add(lint.NewCommand(l, inst.commands))

  // ...

  return inst, nil
}
```
