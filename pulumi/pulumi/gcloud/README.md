# POSH pulumi (azure) provider

## Usage

```go
package plugin

type Plugin struct {
	l        log.Logger
  gcloud   *gcloud.GCloud
  cache    cache.Cache
  kubectl  *kubectl.Kubectl
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
	inst := &Plugin{
		l:        l,
    cache:    cache.MemoryCache{},
		commands: command.Commands{},
	}

  // ...

  inst.op, err := onepassword.New(l, inst.cache));
  if err != nil {
    return nil, errors.Wrap(err, "failed to create onepassword")
  }

  inst.kubectl, err = kubectl.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create kubectl")
  }

  inst.gcloud, err = gcloud.New(l, inst.cache)
  if err != nil {
    return nil, errors.Wrap(err, "failed to create gcloud")
  }

	// ...

  inst.commands.MustAdd(pulumi.NewCommand(l, inst.op, inst.gcloud, inst.cache))

	// ...

	return inst, nil
}
```

### Config

```yaml
## az
pulumi:
  path: .posh/pulumi
  configPath: .posh/config/pulumi
  backends:
      prod:
        location: Germany West Central
        bucket: pulumi-state
        project: xxx
        passphrase:
          account: xxxx
          vault: xxxx
          itemId: xxxx
          field: password
```
