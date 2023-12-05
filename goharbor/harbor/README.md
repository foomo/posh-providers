# POSH harbor provider

## Usage

### Plugin

```go
package main

type Plugin struct {
  l        log.Logger
  harbor   *harbor.Harbor
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  var err error
  inst := &Plugin{
    l:        l,
    commands: command.Commands{},
  }

  // ...

  inst.harbor, err = harbor.New(l)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create harbor")
  }

  // ...

  inst.commands.Add(harbor.New(l, inst.harbor))

  // ...

  return inst, nil
}
```

### Config

```yaml
## Harbor
harbor:
  url: https://harbor.foomo.org
  authUrl: https://harbor.foomo.org/c/oidc/login
  project: "foomo"
```
