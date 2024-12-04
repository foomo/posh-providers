# POSH Golang Migrate provider

## Usage

### Plugin

```go
package main

import (
  // ...
  // import database drivers
  _ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
  // import source drivers
  _ "github.com/golang-migrate/migrate/v4/source/file"
  // ...
)

type Plugin struct {
  l log.Logger
}

func New(l log.Logger) (plugin.Plugin, error) {
  var err error
  inst := &Plugin{
    l:        l,
    commands: command.Commands{},
  }

  // ...

  inst.commands.Add(migrate.NewCommand(l))

  // ...

  return inst, nil
}

```

### Config

```yaml
## Migrate
migrate:
  # https://github.com/golang-migrate/migrate/blob/master/README.md#migration-sources
  databases:
    local: pgx5://postgres:postgres@localhost:5432/admin?sslmode=disable
  # https://github.com/golang-migrate/migrate/blob/master/README.md#migration-sources
  sources:
    default: file://migrations
```
