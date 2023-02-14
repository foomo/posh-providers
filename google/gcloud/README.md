# POSH gcloud provider

## Configuration:

```yaml
gcloud:
  login: true
  configPath: .posh/config/gcloud
  accessTokenPath: .posh/config/gcloud/access_tokens
  environments:
    - name: prod
      project: myproject-123456
      clusters:
        - name: default
          role: admin
          region: europe-west6
```

Using only service account access tokens:

```yaml
gcloud:
  login: false
  configPath: ""
  accessTokenPath: ..posh/config/gcloud
  environments:
    - name: prod
      project: myproject-123456
      clusters:
        - name: default
          role: admin
          region: europe-west6
          accessToken:
            field: 1234564dxtuty3vaaxezex4c7ey
            item: 1234564dxtuty3vaaxezex4c7ey
            vault: 1234564dxtuty3vaaxezex4c7ey
            account: foomo
```

*NOTE: Servce access tokens can optionally be retrieved by OnePassword.*

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
    gcloud.NewCommand(l, provider, inst.kubectl, gcloud.CommandWithOnePassword(inst.op)),
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
