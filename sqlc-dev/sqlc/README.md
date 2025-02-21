# POSH sqlc provider

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
		commands: command.Commands{},
	}

	// ...

  inst.commands.MustAdd(sqlc.NewCommand(l, inst.cache))

	// ...

	return inst, nil
}
```

### Config

To install binary locally, add:

```yaml
sqlc:
  tempDir: .posh/tmp/sqlc
  cacheDirDir: .posh/cache/sqlc
```
### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    - name: sqlc
      tap: foomo/tap/sqlc-dev/sqlc
      version: 1.28.0
```
