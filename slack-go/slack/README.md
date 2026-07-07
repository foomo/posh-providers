# POSH slack provider

Slack integration used by other providers to send notifications (channel messages and webhooks).

This provider has no command of its own — construct it in your plugin and pass it to other commands (e.g. `foomo/squadron`).

## Plugin

```go
package plugin

type Plugin struct {
  l        log.Logger
  op       *onepassword.OnePassword
  slack    *slack.Slack
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    commands: command.Commands{},
  }

  // ...

  inst.op, err = onepassword.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create onepassword")
  }

  inst.slack, err = slack.New(l, inst.op)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create slack")
  }

  // ...

  // pass it to other commands
  inst.commands.MustAdd(squadron.NewCommand(l, inst.cache,
    squadron.CommandWithSlack(inst.slack),
  ))

  // ...

  return inst, nil
}
```

The provider is configured through functional options:

- `slack.WithConfigKey("slack")` — read config from a different key (default `slack`).

### Config

Add this to your `.posh.yml` file:

```yaml
slack:
  # 1Password reference to the Slack API token
  token:
    account: <ACCOUNT>
    vault: <VAULT>
    item: <ITEM>
    field: token
  # Named channels, referenced by id
  channels:
    releases: C0123456789
    general: C9876543210
  # Named incoming webhooks, resolved from 1Password
  webhooks:
    releases:
      account: <ACCOUNT>
      vault: <VAULT>
      item: <ITEM>
      field: webhook
```
