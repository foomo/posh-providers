# POSH squadron provider

Adds a `squadron` command to manage your squadron deployments across clusters and
fleets, wiring in `kubectl`, `1password` and optional Slack notifications.

## Plugin

```go
package plugin

type Plugin struct {
	l        log.Logger
  cache    cache.Cache
  op       *onepassword.OnePassword
  kubectl  *kubectl.Kubectl
	commands command.Commands
  squadron *squadron.Squadron
}

func New(l log.Logger) (plugin.Plugin, error) {
	var err error
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

  inst.kubectl, err = kubectl.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create kubectl")
  }

  inst.squadron, err = squadron.New(l, inst.kubectl)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create squadron")
  }

  inst.commands.MustAdd(squadron.NewCommand(l, inst.squadron, inst.kubectl, inst.op, inst.cache))

	// ...

	return inst, nil
}
```

The command is configured through functional options:

- `squadron.CommandWithName("sq")` — rename the command (default `squadron`).
- `squadron.CommandWithBake(true)` — use `docker buildx bake` instead of `build`.
- `squadron.CommandWithNamespaceFn(fn)` — customize the target namespace per cluster/fleet.
- `squadron.CommandWithSlack(inst.slack)` — enable `--slack` deployment notifications.
- `squadron.CommandWithSlackChannelID("squadron")` — Slack channel to notify (default `squadron`).
- `squadron.CommandWithSlackWebhookID("...")` — Slack webhook to notify instead of a channel.

The provider reads its config from the `squadron` key by default; override it with
`squadron.WithConfigKey("...")`.

### Config

```yaml
squadron:
  # Path to the squadron root
  path: squadrons
  # Cluster configurations
  clusters:
    - name: prod
      # Enable slack notification by default
      notify: true
      # Require interactive confirmation for up/down/rollback
      confirm: true
      # Cluster fleet names
      fleets: ["default"]
    - name: dev
      fleets: ["default"]
```
