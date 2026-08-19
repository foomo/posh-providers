# POSH golang provider

## Usage

Register it with the plugin's command set:

```go
import (
	"github.com/foomo/posh-providers/golang"
)

inst.commands.Add(golang.NewCommand(l, c))
```

The command is registered under the name `go`.

The provider takes no configuration. Build tags come from the `GO_BUILD_TAGS` environment variable
and default to `safe`.

`golang.CommandWithExecGolangciLint` overrides how the `golangci-lint` binary is invoked, e.g. to
point at an ownbrew-managed copy.

`Command` implements `arbitrary/lint`'s `Linter` interface, so registering it also enrols
`golangci-lint` in the project-wide `lint` command.
