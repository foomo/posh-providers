# POSH az provider

## Usage

This provider requires `az` to be installed on your system.

### Plugin

```go
inst.commands.Add(az.NewCommand(l, inst.az, inst.kubectl))
```

### Config

```yaml
az:
  configPath: .posh/config/azure
  tenantId: xxxx-xx-xx-xx-xxxx
  subscriptions:
    development:
      name: xxxx-xx-xx-xx-xxxx
      clusters:
        dev:
          name: my-cluster
          resourceGroup: my-resource-group
      artifactories:
        dev:
          name: my-artifactory
          resourceGroup: my-resource-group
```

### Commands

```shell
> help az
Manage azure resources

Usage:
      az [command]

Available Commands:
      login                         Log in to Azure
      logout                        Log out to remove access to Azure subscriptions
      configure                     Manage Azure CLI configuration
      artifactory                   Login into the artifactory
      kubeconfig                    Retrieve credentials to access remote cluster
```

#### Examples

```shell
# Log into azure tenant
> az login

# Authorize artifactory
> az artifactory <SUBSCRIPTION> <ARTIFACTORY>

# Retrieve cluster kubeconfig
> az kubeconfig <SUBSCRIPTION> <CLUSTER>
```
