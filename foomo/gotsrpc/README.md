# POSH gotsrpc provider

## Usage

### Plugin

```go
package plugin

import (
	"github.com/foomo/posh/provider/foomo/gotsrpc"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/plugin"
	"github.com/spf13/viper"
)

type Plugin struct {
	l        log.Logger
	commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
	inst := &Plugin{
		l:        l,
		commands: command.Commands{},
	}

	// ...

  inst.commands.Add(gotsrpc.NewCommand(l))

	// ...

	return inst, nil
}
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    - name: gotsrpc
      tap: foomo/tap/foomo/gotsrpc
      version: 2.6.2
```
