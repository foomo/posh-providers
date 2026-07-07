# POSH webdriverio provider

Adds a `wdio` command that runs [WebdriverIO](https://webdriver.io/) end-to-end
tests across your configured modes, sites and environments, resolving basic auth
and BrowserStack credentials from 1Password.

## Usage

```
Run wdio commands

Usage:
  wdio <mode> <site> <env> [path]
```

The `mode`, `site` and `env` arguments are suggested from your configuration.
The optional `path` argument narrows the run to a single `wdio.conf.ts`
directory. Flags include `--spec`, `--suite`, `--ci`, `--headless`, `--debug`
and `--bail`.

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

  // add command
  cmd, err := webdriverio.NewCommand(l, inst.cache, inst.op)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create wdio command")
  }
  inst.commands.MustAdd(cmd)

  // ...

  return inst, nil
}
```

The command is configured through functional options:

- `webdriverio.CommandWithName("e2e")` — rename the command (default `wdio`).
- `webdriverio.WithConfigKey("e2e")` — read config from a different key (default `webdriverio`).

### Config

```yaml
webdriverio:
  # Directories containing e2e/wdio.conf.ts
  dirs:
    - ./frontend
  # Run modes, suggested as the first argument
  modes:
    local:
      port: "8443"
      hostPrefix: local
  # Sites and their environments, suggested as the second and third arguments
  sites:
    shop:
      staging:
        domain: staging.example.com
        auth:
          item: <document>
          vault: <vault>
          account: <account>
  # Named 1Password secrets
  secrets:
    api:
      item: <document>
      vault: <vault>
      account: <account>
  # BrowserStack credentials (username/password fields)
  browserStack:
    item: <document>
    vault: <vault>
    account: <account>
```
