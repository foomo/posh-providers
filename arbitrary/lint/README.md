# POSH lint provider

> Adds configured linters to your project.

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

  inst.commands.Add(lint.NewCommand(l, inst.cache,
    lint.CommandWithGo(),
    lint.CommandWithTSC(),
    lint.CommandWithHelm(),
    lint.CommandWithESLint(),
    lint.CommandWithGherkin(),
    lint.CommandWithTerraform(),
    lint.CommandWithTerrascan(),
	))

  // ...

  return inst, nil
}
```
