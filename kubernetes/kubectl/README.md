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

### Command

Register the `kubectl` command to run kubectl against a cluster:

```go
inst.commands.Add(kubectl.NewCommand(l, inst.kubectl))
```

Usage — the `<cluster>` argument selects the kubeconfig, the optional `--profile` flag selects
a profile, and all remaining args and flags are passed straight through to the `kubectl` CLI:

```shell
kubectl <cluster> [--profile <profile>] get pods -n <namespace>
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
