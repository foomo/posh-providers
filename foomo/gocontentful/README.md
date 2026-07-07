# POSH gocontentful provider

Adds a `gocontentful` command that generates type-safe Contentful client code by
running [gocontentful](https://github.com/foomo/gocontentful) for every
`gocontentful.yaml` found in your project.

## Plugin

```go
package plugin

type Plugin struct {
  l        log.Logger
  op       *onepassword.OnePassword
  cache    cache.Cache
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    cache:    cache.MemoryCache{},
    commands: command.Commands{},
  }

  // ...

  inst.op, err = onepassword.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create onepassword")
  }

  // ...

  inst.commands.MustAdd(gocontentful.NewCommand(l, inst.cache, inst.op))

  // ...

  return inst, nil
}
```

The command is configured through functional options:

- `gocontentful.CommandWithName("contentful")` — rename the command (default `gocontentful`).

### Config

```yaml
gocontentful:
  # Contentful space id
  spaceId: <space-id>
  # Contentful management API key
  cmaKey: <cma-key>
  # Contentful environment; defaults to "master" when empty
  environment: master
  # Content types to generate code for
  contentTypes:
    - foo
    - bar
```
