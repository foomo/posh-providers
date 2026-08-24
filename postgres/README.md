# POSH postgres provider

## Usage

### Plugin

```go
package plugin

type Plugin struct {
	l        log.Logger
	zip      *zip.Zip
	commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
	var err error

	inst := &Plugin{
		l:        l,
		commands: command.Commands{},
	}

	// ...

	// Required for `dump --zip` / `--zip-cred`; without it those fail with a
	// message naming this option. The other verbs work regardless.
	inst.zip, err = zip.New(l, inst.onePassword)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create zip")
	}

	// ...

  inst.commands.Add(postgres.NewCommand(l, postgres.CommandWithZip(inst.zip)))

	// ...

	return inst, nil
}
```

### Dependencies

This requires you to have:

- psql
- pg_dump
- pg_restore
