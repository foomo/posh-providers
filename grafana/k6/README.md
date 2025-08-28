# POSH k6 provider

## Usage

### Plugin

```go
package main

type Plugin struct {
  l        log.Logger
  op       *onepassword.OnePassword
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    commands: command.Commands{},
  }

  // ...

  if value, err := onepassword.New(l, inst.cache); err != nil {
    return nil, err
  } else {
    inst.op = value
  }

  // ...

  inst.commands.MustAdd(k6.NewCommand(l, inst.op))

  // ...

  return inst, nil
}
```

### Config

```yaml
k6:
  path: ./devtools/k6
  envs:
    dev:
      URL: https://quickpizza.grafana.com
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    # https://github.com/grafana/k6/releases
    - name: k6
      tap: foomo/tap/grafana/k6
      version: 1.2.2
```

