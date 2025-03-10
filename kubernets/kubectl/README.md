# POSH kubectl provider

## Usage

### Plugin

```go
package main

type Plugin struct {
  l        log.Logger
  cache    cache.Cache
  kubectl  *kubectl.Kubectl
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

  inst.kubectl, err = kubectl.New(l, inst.c)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create kubectl")
  }

  // ...

  return inst, nil
}
```

### Config

```yaml
## kubectl
kubectl:
  configPath: .posh/config/kubectl
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    ## https://kubernetes.io/releases/
    - name: kubectl
      tap: foomo/tap/kubernetes/kubectl
      version: 1.28.4
```
