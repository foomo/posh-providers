# POSH Helm provider

## Usage

```go
func New(l log.Logger) (plugin.Plugin, error) {
	// ...
  inst.commands.Add(helm.NewCommand(l, kubectl))
	// ...
}
```
