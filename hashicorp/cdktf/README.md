# POSH CDKTF provider

## Config

```yaml
cdktf:
  path: devops/cdktf
```

## Usage

```go
func New(l log.Logger) (plugin.Plugin, error) {
	// ...
  inst.commands.Add(helm.NewCommand(l, kubectl))
	// ...
}
```
