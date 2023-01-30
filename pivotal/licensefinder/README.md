# POSH License Finder provider

## Config:

```yaml
licenseFinder:
  logPath: .posh/logs/licenses.log
  decisionsPath: .posh/licenses.yaml
```

## Usage

```go
func New(l log.Logger) (plugin.Plugin, error) {
	// ...
  inst.commands.Add(licensefinder.NewCommand(l, cache))
	// ...
}
```
