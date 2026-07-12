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
      keyVaults:
        dev:
          name: my-key-vault
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
      vault                         Manage key vault entries
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

#### Key vault

Manage keys, secrets and certificates of a configured key vault. Every command is scoped by
`<SUBSCRIPTION>` and `<VAULT>` (both tab-completable from the config), followed by the entry type
(`key`, `secret`, `certificate`) and the operation (`set`, `list`, `delete`).

```shell
# Create a new key (or key version)
> az vault <SUBSCRIPTION> <VAULT> key set <NAME> --kty RSA --size 2048
# List keys
> az vault <SUBSCRIPTION> <VAULT> key list
# Delete a key (name tab-completes from the vault)
> az vault <SUBSCRIPTION> <VAULT> key delete <NAME>

# Create or update a secret
> az vault <SUBSCRIPTION> <VAULT> secret set <NAME> --value <VALUE>
> az vault <SUBSCRIPTION> <VAULT> secret set <NAME> --file ./secret.txt
# List / delete secrets
> az vault <SUBSCRIPTION> <VAULT> secret list
> az vault <SUBSCRIPTION> <VAULT> secret delete <NAME>

# Import a certificate from a PEM or PFX file
> az vault <SUBSCRIPTION> <VAULT> certificate set <NAME> --file ./cert.pfx --password <PASSWORD>
# List / delete certificates
> az vault <SUBSCRIPTION> <VAULT> certificate list
> az vault <SUBSCRIPTION> <VAULT> certificate delete <NAME>
```
