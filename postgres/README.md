# POSH gotsrpc provider

## Usage

### Plugin

```go
package plugin

type Plugin struct {
	l        log.Logger
	commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
	inst := &Plugin{
		l:        l,
		commands: command.Commands{},
	}

	// ...

  inst.commands.Add(postgres.NewCommand(l))

	// ...

	return inst, nil
}
```

### Dependencies

This requires you to have:

- psql
- pg_dump
- pg_restore
