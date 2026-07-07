# POSH yarn provider

Adds a `yarn` command to run yarn commands within your project.

## Usage

The command exposes the following subcommands:

- `install [path]` — install dependencies (optionally in the given path).
- `run [path] <script>` — run a script, suggested from the `package.json` scripts.
- `run-all <script>` — run a script in all discovered `package.json` directories, with an optional `--parallel` limit.

Any other arguments are passed straight through to `yarn`.

## Plugin

```go
package main

type Plugin struct {
  l        log.Logger
  cache    cache.Cache
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  var err error
  inst := &Plugin{
    l:        l,
    cache:    &cache.MemoryCache{},
    commands: command.Commands{},
  }

  // ...

  inst.commands.MustAdd(yarn.NewCommand(l, inst.cache))

  // ...

  return inst, nil
}
```

The command is configured through functional options:

- `yarn.CommandWithName("yarn")` — rename the command (default `yarn`).
