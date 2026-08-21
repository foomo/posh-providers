# POSH gotsrpc provider

## Usage

### Plugin

```go
package plugin

type Plugin struct {
	l        log.Logger
	cache    cache.Cache
	commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
	inst := &Plugin{
		l:        l,
		cache:    &cache.MemoryCache{},
		commands: command.Commands{},
	}

	// ...

  inst.commands.Add(gotsrpc.NewCommand(l, inst.cache))

	// ...

	return inst, nil
}
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    - name: gotsrpc
      tap: foomo/tap/foomo/gotsrpc
      version: 2.6.2
```
