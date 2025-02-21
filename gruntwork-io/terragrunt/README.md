# POSH terragrunt provider

## Usage

### Plugin

```go
func New(l log.Logger) (plugin.Plugin, error) {
	// ...
  inst.commands.MustAdd(terragrunt.NewCommand(l, inst.op, inst.cache))
	// ...
}
```

### Config

```yaml
terragrunt:
  path: path/to/terragrunt
  cachPath: .posh/cache/terragrunt
```
