# POSH sesamy provider

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

  inst.commands.MustAdd(sesamy.NewCommand(l, inst.onePassword))

	// ...

	return inst, nil
}
```

### Config

```yaml
sesamy:
  default:
    - path/to/sesamy.base.yaml
    - path/to/sesamy.base.override.yaml
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    - name: sesamy
      tap: foomo/tap/foomo/sesamy
      version: 1.0.0
```
