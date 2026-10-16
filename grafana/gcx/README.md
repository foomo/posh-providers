# POSH gcx provider

Wraps the [Grafana CLI](https://github.com/grafana/gcx) (`gcx`). The first argument selects the environment, which maps
to a gcx config file (containing a single context) inside the configured path — everything after it is passed through to
`gcx` and completed by gcx's own shell completion.

```shell
> gcx dev dashboards list
# runs: GCX_CONFIG=./devops/gcx/dev.yaml gcx dashboards list
```

## Usage

### Plugin

```go
package main

type Plugin struct {
  l        log.Logger
  cache    cache.Cache
  kubectl  *kubectl.Kubectl
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    cache:    cache.NewMemoryCache(),
    commands: command.Commands{},
  }

  // ...

  if value, err := gcx.NewCommand(l, inst.cache, gcx.CommandWithKubectl(inst.kubectl)); err != nil {
    return nil, err
  } else {
    inst.commands.MustAdd(value)
  }

  // ...

  return inst, nil
}
```

`CommandWithKubectl` is only required for environments that set `urlForward`; omit it otherwise.

### Config

```yaml
gcx:
  path: ./devops/gcx
```

Every `*.yaml` / `*.yml` file in `path` is an environment, e.g. `./devops/gcx/dev.yaml` becomes `gcx dev ...`.

An environment can set `urlForward` to reach a Grafana server that is only reachable from inside a
cluster. When set, `GRAFANA_SERVER` is the local address of a `kubectl port-forward` to the given
resource instead of `url`. `cluster` selects the kubectl cluster (defaults to the environment name); `cluster` and `profile` are optional:

```yaml
urlForward:
  resource: svc/grafana
  namespace: monitoring
  port: 3000
  cluster: ""
  profile: ""
```

### Install

```shell
brew install grafana/grafana/gcx
```

## Notes

- Completions are cached per command path (not per environment) since the gcx command tree is identical for all
  environments. Run `cache clear` in posh to flush them.
- Flags are not completed, they are passed through verbatim.
