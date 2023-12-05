# POSH Zeus provider

## Usage

### Plugin

```go
package plugin

type Plugin struct {
	l        log.Logger
	c        cache.Cache
	commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
	inst := &Plugin{
		l:        l,
		c:        cache.MemoryCache{},
		commands: command.Commands{},
	}

	// ...

  inst.commands.Add(zeus.NewCommand(l, inst.c))

	// ...

	return inst, nil
}
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    - name: zeus
      tap: foomo/tap/dreadl0ck/zeus
      version: 0.9.11
```
