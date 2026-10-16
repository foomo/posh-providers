# gnat

> [gnat](https://github.com/galaxy-io/gnat) — a NATS JetStream terminal UI client.

Holds a color theme and a map of name → NATS server profile. The names autocomplete in the shell
(with their description); the selected name resolves to its profile and its URL is passed to the
`gnat` CLI as `-url`, along with `-theme`.

A profile can set `urlForward` instead of `url` to reach a server only reachable from inside a
cluster: a `kubectl port-forward` to the given resource is opened and its local address
(`127.0.0.1:<port>`) is passed as `-url`. `urlForward.cluster` selects the kubectl cluster (defaults to the profile name), and
the command must be created with `gnat.CommandWithKubectl(...)`.

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

  cmd, err := gnat.NewCommand(l, gnat.CommandWithKubectl(inst.kubectl)) // kubectl only needed for urlForward
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
    cluster:
      urlForward:
        resource: svc/nats
        namespace: gnat
        port: 4222
        cluster: ""
        profile: ""
      description: In-cluster server via port-forward
```
