# POSH License Finder provider

## Usage

### Plugin

```go
func New(l log.Logger) (plugin.Plugin, error) {
	// ...
  inst.commands.Add(licensefinder.NewCommand(l, cache))
	// ...
}
```

## Config

```yaml
licenseFinder:
  logPath: .posh/logs/licenses.log
  decisionsPath: .posh/licenses.yaml
```

