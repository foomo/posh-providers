# POSH beam provider

## Usage

### Plugin

```go
package plugin

type Plugin struct {
	l           log.Logger
  beam        *beam.Beam
  cloudflared *cloudflared.cloudflared
  cache       cache.Cache
	commands    command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
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

  inst.cloudflared, err = cloudflared.New(l)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create cloudflared")
  }

  inst.beam, err = beam.NewBeam(l, inst.op)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create beam")
  }

  // ...
  inst.commands.Add(command.NewCheck(l,
    beam.ClusterChecker(inst.cloudflared, inst.beam.Config().GetCluster("my-cluster")),
    beam.DatabaseChecker(inst.cloudflared, inst.beam.Config().GetDatabase("my-database")),
  ))

  inst.commands.MustAdd(beam.NewCommand(l, inst.beam, inst.kubectl, inst.cloudflared))

	// ...

	return inst, nil
}
```

### Config

```yaml
beam:
  clusters:
    my-cluster:
      port: 12200
      hostname: "my-concierge.domain.com"
      kubeconfig:
        item: <document>
        vault: <vault>
        account: <account>
  databases:
    my-database:
      port: 12202
      hostname: "my-database.domain.com"
```
