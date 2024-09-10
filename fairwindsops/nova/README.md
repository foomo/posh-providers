# POSH nova provider

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

  inst.commands.Add(nova.NewCommand(l, inst.kubectl))

  // ...

  return inst, nil
}
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    # https://github.com/FairwindsOps/nova/releases
    - name: nova
      tap: foomo/tap/fairwindsops/nova
      version: 3.10.1
```
