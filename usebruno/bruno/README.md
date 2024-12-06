# POSH Bruno provider

## Usage

### Plugin

```go
package main

import (
  "github.com/foomo/posh-providers/onepassword"
  "github.com/foomo/posh-providers/usebruno/bruno"
  "github.com/foomo/posh/pkg/cache"
  "github.com/foomo/posh/pkg/command"
)

type Plugin struct {
  l     log.Logger
  cache cache.Cache
  op    *onepassword.OnePassword
}

func New(l log.Logger) (plugin.Plugin, error) {
  var err error
  inst := &Plugin{
    l:        l,
    cache:    cache.MemoryCache{},
    commands: command.Commands{},
  }

  // ...

  inst.op, err = onepassword.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create onepassword")
  }

  // ...

  inst.commands.Add(bruno.NewCommand(l, bruno.CommandWithOnePassword(inst.op)))

  // ...

  return inst, nil
}

```

### Config

```yaml
## Bruno
bruno:
  path: '${PROJECT_ROOT}/.posh/bruno'
```

### OnePassword

To inject secrets from 1Password, create a `bruno.env` file:

```text
JWT_TOKEN=*********************
```

Render the file to `.env`:

```shell
> bruno env
```

And use the secret in your environment `environments/local.bru`:

```text
vars {
  host: http://localhost:5005
  jwtToken: {{process.env.JWT_TOKEN}}
}
```
