# tempo

> [tempo](https://github.com/galaxy-io/tempo) — a Temporal workflow terminal UI client.

Holds a color theme and a map of name → Temporal connection profile. The names autocomplete in the
shell (with their description); the selected name resolves to its profile and its URL is passed to the
`tempo` CLI as `--address`, along with `--theme`.

A profile can set `urlForward` instead of `url` to reach a server only reachable from inside a
cluster: a `kubectl port-forward` to the given resource is opened and its local address
(`127.0.0.1:<port>`) is passed as `--address`. `urlForward.cluster` selects the kubectl cluster (defaults to the profile name), and
the command must be created with `tempo.CommandWithKubectl(...)`.

An optional `configDir` sets `$XDG_CONFIG_HOME` for the `tempo` process, so its configuration files
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

  cmd, err := tempo.NewCommand(l, tempo.CommandWithKubectl(inst.kubectl)) // kubectl only needed for urlForward
  if err != nil {
    return nil, errors.Wrap(err, "failed to create tempo command")
  }

  inst.commands.Add(cmd)

  return inst, nil
}
```

## Configuration

```yaml
## tempo
tempo:
  theme: tokyonight-night
  configDir: .posh/config/tempo
  profiles:
    local:
      url: localhost:7233
      description: Local dev server
    prod:
      url: temporal.prod.example.com:7233
      description: Production
    cluster:
      urlForward:
        resource: svc/temporal-frontend
        namespace: temporal
        port: 7233
        cluster: ""
        profile: ""
      description: In-cluster server via port-forward
```
