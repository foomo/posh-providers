# POSH mjml provider

Adds an `mjml` command that compiles your `*.mjml` templates to HTML.

## Usage

```
Run mjml

Usage:
  mjml [path]
```

The `path` argument is optional and defaults to the current directory; it is
suggested from directories containing `*.mjml` sources. Use the `--parallel`
flag to set the number of parallel compile processes (default `0`, i.e. one at
a time).

## Plugin

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

  inst.commands.MustAdd(mjml.NewCommand(l, inst.cache))

  // ...

  return inst, nil
}
```

The command is configured through functional options:

- `mjml.CommandWithName("email")` — rename the command (default `mjml`).
