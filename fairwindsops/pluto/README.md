# POSH pluto provider

## Usage

### Plugin

```go
package main

type Plugin struct {
  l        log.Logger
  kubectl  *kubectl.Kubectl
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  var err error
  inst := &Plugin{
    l:        l,
    commands: command.Commands{},
  }

  // ...

  inst.kubectl, err = kubectl.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create kubectl")
  }


  // ...

  inst.commands.Add(pluto.NewCommand(l, inst.kubectl))

  // ...

  return inst, nil
}
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    # https://github.com/FairwindsOps/pluto/releases
    - name: pluto
      tap: foomo/tap/fairwindsops/pluto
      version: 5.20.2
```
