# gnat

> [gnat](https://github.com/galaxy-io/gnat) — a NATS JetStream terminal UI client.

Holds a color theme and a map of name → NATS server URL. The names autocomplete in the shell; the
selected name resolves to its URL and is passed to the `gnat` CLI as `-url`, along with `-theme`.

## Usage

```go
package main

type Plugin struct {
  l        log.Logger
  cache    cache.Cache
  gnat     *gnat.GNAT
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  var err error
  inst := &Plugin{
    l:        l,
    cache:    &cache.MemoryCache{},
    commands: command.Commands{},
  }

  inst.gnat, err = gnat.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create gnat")
  }

  inst.commands.Add(gnat.NewCommand(l, inst.cache, inst.gnat))

  return inst, nil
}
```

## Configuration

```yaml
## gnat
gnat:
  theme: tokyonight-night
  urls:
    local: nats://localhost:4222
    prod: nats://nats.prod.example.com:4222
```
