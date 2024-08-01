# POSH beam provider

## Usage

### Plugin

```go
package plugin

type Plugin struct {
	l        log.Logger
  beam     *beam.Beam
  cache    cache.Cache
	commands command.Commands
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

  inst.beam, err = beam.NewBeam(l, inst.op)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create beam")
  }

  // ...
  inst.commands.Add(command.NewCheck(l,
    beam.TunnelChecker(inst.beam, "my-env", "my-cluster"),
  ))

  inst.commands.MustAdd(beam.NewCommand(l, inst.beam, inst.kubectl))

	// ...

	return inst, nil
}
```

### Config

```yaml
beam:
  my-env:
    clusters:
      my-cluster:
        port: 1234
        hostname: beam.my-domain.com
        credentials:
          item: <name|uuid>
          vault: <name|uuid>
          account: <account>
```

### Ownbrew

```yaml
ownbrew:
  packages:
    - name: cloudflared
      tap: foomo/tap/cloudflare/cloudflared
      version: 2024.6.1
```
