# K3d

## Usage

### Plugin

```go
package plugin

type Plugin struct {
	l        log.Logger
  k3d      *k3d.K3d
	commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
	inst := &Plugin{
		l:        l,
		commands: command.Commands{},
	}

  var err error

  // ...

  inst.k3d, err = k3d.New(l)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create k3d")
  }

	// ...

  inst.commands.MustAdd(k3d.NewCommand(l, inst.k3d, inst.kubectl))

	// ...

	return inst, nil
}
```

### Config

```yaml
## K3d
k3d:
  charts:
    path: devops/k3d
    prefix: shared-
  registry:
    name: foomo-registry
    port: 12345
  clusters:
    - name: local
      port: 9443
      alias: foomo
      image: rancher/k3s:v1.28.2-k3s1
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    ## https://github.com/k3d-io/k3d/releases
    - name: k3d
      tap: foomo/tap/k3d-io/k3d
      version: 5.6.0
```
