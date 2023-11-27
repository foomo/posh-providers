# POSH open provider

> Define and open common URL from the posh.

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
  op       *onepassword.OnePassword
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

  inst.op, err = onepassword.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create onepassword")
  }

	// ...

  inst.commands.MustAdd(open.NewCommand(l, inst.op))

  // ...

  return inst, nil
}
```

### Config

```yaml
## Open
open:
  homepage:
    description: Home page
    routes:
      home:
        path: https://www.foomo.org/
        description: Home
      imprint:
        path: https://www.foomo.org/imprint
        description: Imprint
        basicAuth:
          item: xxxxxxxxxxxxxxxxxxxxxxxxxx
          vault: xxxxxxxxxxxxxxxxxxxxxxxxxx
          account: foomo
```

