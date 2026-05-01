# POSH terraform provider

## Usage

### Plugin

```go
func New(l log.Logger) (plugin.Plugin, error) {
	// ...
  inst.commands.MustAdd(terraform.NewCommand(l, inst.cache))
	// ...
}
```

### Config

```yaml
terraform:
  path: path/to/terraform/workspaces
  subscriptions:
    my-workspace:
      id: 00000000-0000-0000-0000-000000000000
      backend:
        resourceGroupName: my-rg
        storageAccountName: mystorageaccount
        containerName: tfstate
  servicePrincipals:
    my-sp:
      tenantId: 00000000-0000-0000-0000-000000000000
      clientId: 00000000-0000-0000-0000-000000000000
      clientSecret: my-secret
      subscriptionId: 00000000-0000-0000-0000-000000000000
```
