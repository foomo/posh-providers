# POSH squadron provider

## Usage

### Plugin

```go
package plugin

type Plugin struct {
	l        log.Logger
  cache    cache.Cache
  op       *onepassword.OnePassword
  kubectl  *kubectl.Kubectl
	commands command.Commands
  squadron *squadron.Squadron
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

  inst.kubectl, err = kubectl.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create kubectl")
  }

  inst.squadron, err = squadron.New(l, inst.kubectl)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create squadron")
  }

  inst.commands.Add(squadron.NewCommand(l, inst.squadron, inst.kubectl, inst.op, inst.cache))

	// ...

	return inst, nil
}
```

### Config

To install binary locally, add:

```yaml
squadron:
  path: squadrons
  clusters:
    - name: prod
      notify: true
      confirm: true
      fleets: ["default"]
    - name: dev
      fleets: ["default"]
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    # https://github.com/foomo/squadron/releases
    - name: squadron
      tap: foomo/tap/foomo/squadron
      version: 2.1.5
```
