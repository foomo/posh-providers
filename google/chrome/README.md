# POSH Google Chrome provider

Adds a `chrome` command that opens Google Chrome with an isolated user-data
directory per profile, so you can keep separate browser environments (URLs,
proxies, incognito) for your project.

## Usage

```
Open a Google Chrome browser

Usage:
  chrome <profile> [url]
```

The `profile` argument is required and is suggested from your configured
profiles. The `url` argument is optional; when given it overrides the profile's
default `url`.

## Plugin

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

  // add command
  cmd, err := chrome.NewCommand(l)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create chrome command")
  }
  inst.commands.Add(cmd)

  // ...
}
```

The command is configured through functional options:

- `chrome.CommandWithName("browser")` — rename the command (default `chrome`).
- `chrome.CommandWithConfigKey("browser")` — read config from a different key (default `chrome`).

### Config

```yaml
chrome:
  # Directory for per-profile browser data dirs; created on startup
  path: .posh/browser
  # Browser binary; defaults to Chrome on macOS/Windows, `google-chrome` on Linux
  command: ""
  # Open every profile in incognito mode
  incognito: false
  # Named environments, suggested as the first argument
  profiles:
    staging:
      # Launch in app mode (--app=<url>) instead of a normal browser window
      app: true
      # Default URL to open when no url argument is given
      url: https://staging.example.com
      # Proxy server passed to Chrome via --proxy-server
      proxy: socks5://localhost:1080
      # Open this profile in incognito mode
      incognito: true
    prod:
      url: https://example.com
```
