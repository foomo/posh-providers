# POSH gcloud provider

## Configuration:

```yaml
gcloud:
  configPath: .posh/config/gcloud
  clusters:
    prod:
      name: default
      project: myproject-123456
      region: europe-west6
```

Using service account access tokens:

```yaml
gcloud:
  configPath: .posh/config/gcloud
  accounts:
    prod:
      name: user@account.iam.gserviceaccount.com
    admin@prod:
      name: admin@account.iam.gserviceaccount.com
  clusters:
    prod:
      name: default
      project: myproject-123456
      region: europe-west6
      account: prod
    admin@prod:
      name: default
      project: myproject-123456
      region: europe-west6
      account: admin@prod
```

*NOTE: Servce account keys can optionally be retrieved by OnePassword.*

```yaml
gcloud:
  configPath: .posh/config/gcloud
  accounts:
    prod:
      name: user@account.iam.gserviceaccount.com
      key:
        field: 1234564dxtuty3vaaxezex4c7ey
        item: 1234564dxtuty3vaaxezex4c7ey
        vault: 1234564dxtuty3vaaxezex4c7ey
        account: foomo
```

## Usage

```go
func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{}

	// ...

  // create provider
  provider, err := gcloud.New(l, inst.cache)
  if err != nil {
    return nil
  }

  // add command
  inst.commands.Add(
    gcloud.NewCommand(l, provider, inst.kubectl),
	)

	// ...
}
```

Using service account access tokens retrieved by OnePassword:

```go
func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{}

  // ...

  // create provider
  provider, err := gcloud.New(l, inst.cache)
  if err != nil {
    return nil
  }

  // add command
  inst.commands.Add(
    gcloud.NewCommand(l, provider, inst.kubectl, gcloud.CommandWithOnePassword(inst.op)),
	)

	// ...
}
```

## Ownbrew

```yaml
require:
  packages:
    - name: gcloud
      version: '>=409'
      command: gcloud --version 2>&1 | grep "Google Cloud SDK" | awk '{print $4}'
      help: |
        Please ensure you have 'gcloud' installed in a recent version: %s!

          $ brew update
          $ brew install google-cloud-sdk
```
