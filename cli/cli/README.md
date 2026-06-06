# POSH gh provider

Wraps the [GitHub CLI (`gh`)](https://github.com/cli/cli) for use in a posh shell. Currently exposes `gh auth status` and `gh auth refresh` (with `--reset-token` and `--scopes`).

## Usage

### Plugin

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

  // ...

  ghCLI, err := cli.New(l)
  if err != nil {
    return nil, err
  }

  ghCmd, err := cli.NewCommand(l, ghCLI)
  if err != nil {
    return nil, err
  }

  inst.commands.MustAdd(ghCmd)
  inst.checkers.MustAdd(cli.ScopesChecker(ghCLI))

  // ...

  return inst, nil
}
```

### Config

```yaml
## gh
gh:
  scopes:
    github.com:
      - repo
      - read:packages
```

`scopes` declares required OAuth scopes per gh hostname. The `ScopesChecker` runs each prompt tick, calls `gh auth status --hostname <host>`, and reports any missing scopes.

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    ## https://github.com/cli/cli/releases
    - name: gh
      tap: foomo/cli/cli
      version: 2.65.0
```

### Commands

```shell
# View authentication status
> gh auth status

# Refresh authentication, resetting the stored token
> gh auth refresh --reset-token

# Refresh authentication and request additional scopes
> gh auth refresh --scopes repo --scopes workflow
```
