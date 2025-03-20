# POSH az provider

## Usage

This provider requires `az` to be installed on your system.

### Plugin

```go
package main

type Plugin struct {
  l        log.Logger
  az        *az.AZ
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

  inst.kubectl, err = kubectl.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create kubectl")
  }

  inst.az, err = az.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create az")
  }

  // ...

  inst.commands.Add(az.NewCommand(l, inst.az, inst.kubectl))

  // ...

  return inst, nil
}
```

### Config

```yaml
## az
az:
  configPath: .posh/config/azure
  subscriptions:
    production:
      name: my-subscription
      clusters:
        default:
          name: aks-my-prod
      artifactories:
        default:
          name: acr-my-prod
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    ## https://github.com/Azure/kubelogin/releases
    - name: kubelogin
      tap: foomo/tap/azure/kubelogin
      version: 0.1.0
```
