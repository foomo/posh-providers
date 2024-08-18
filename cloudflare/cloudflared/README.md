# POSH cloudflared provider

## Usage

### Plugin

```go
package plugin

type Plugin struct {
  l           log.Logger
  cloudflared *cloudflared.Cloudflared
  commands    command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    commands: command.Commands{},
  }

  // ...

  inst.cloudflared, err = cloudflared.New(l)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create cloudflared")
  }

  // ...
  inst.commands.Add(command.NewCheck(l,
    cloudflared.AcccessChecker(inst.cloudflared, ints.cloudflared.Config().GetAccess("my-access")),
  ))

  inst.commands.MustAdd(cloudflared.NewCommand(l, inst.cloudflared))

  // ...

  return inst, nil
}

```

### Config

```yaml
cloudflared:
  path: devops/config/cloudflared
  access:
    my-access:
      type: tcp
      port: 1234
      hostname: cloudflared.my-domain.com
```

### Ownbrew

```yaml
ownbrew:
  packages:
    - name: cloudflared
      tap: foomo/tap/cloudflare/cloudflared
      version: 2024.6.1
```
