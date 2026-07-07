# gnat

> [gnat](https://github.com/galaxy-io/gnat) — a NATS JetStream terminal UI client.

Holds a color theme and a map of name → NATS server profile. The names autocomplete in the shell
(with their description); the selected name resolves to its profile and its URL is passed to the
`gnat` CLI as `-url`, along with `-theme`.

An optional `configDir` sets `$XDG_CONFIG_HOME` for the `gnat` process, so its configuration files
can be kept alongside the project instead of the user's home directory.

## Usage

```go
package main

type Plugin struct {
  l        log.Logger
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    commands: command.Commands{},
  }

  cmd, err := gnat.NewCommand(l)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create gnat command")
  }

  inst.commands.Add(cmd)

  return inst, nil
}
```

## Configuration

```yaml
## gnat
gnat:
  theme: tokyonight-night
  configDir: .posh/config/gnat
  profiles:
    local:
      url: nats://localhost:4222
      description: Local dev server
    prod:
      url: nats://nats.prod.example.com:4222
      description: Production
```
