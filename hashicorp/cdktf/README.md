# POSH CDKTF provider

## Usage

### Plugin

```go
func New(l log.Logger) (plugin.Plugin, error) {
	// ...
  inst.commands.Add(helm.NewCommand(l, kubectl))
	// ...
}
```

### Config

```yaml
cdktf:
  path: devops/cdktf
```
