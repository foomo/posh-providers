# POSH 1Password provider

## Configuration

Add this to your '.posh.yml' file:

```yaml
onePassword:
  account: <ACCOUNT>
  tokenFilename: .posh/config/.op
```

## Usage:

```go
package plugin

import (
	"github.com/foomo/posh/provider/onepassword"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/plugin"
	"github.com/spf13/viper"
)

type Plugin struct {
	l        log.Logger
	c        cache.Cache
	commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
	inst := &Plugin{
		l:        l,
		c:        cache.MemoryCache{},
		commands: command.Commands{},
	}

	// ...

	// 1Password
	if onePassword, err := onepassword.New(l, inst.c, onepassword.WithTokenFilename(viper.GetString("onePassword.tokenFilename"))); err != nil {
		return nil, err
	} else if cmd, err := onepassword.NewCommand(l, onePassword); err != nil {
		return nil, err
	} else {
		inst.commands.Add(cmd)
	}

	// ...

	return inst, nil
}
```

## Require

To add a requirement check for op, add:

```yaml
require:
  scripts:
    - name: op
      command: |
        [[ $(op account --account <ACCOUNT> get 2>&1) =~ "found no account" ]] && exit 1 || exit 0
      help: |
        You're 1Password account is not registered yet! Please do so by running:

          $ op account add --address <ACCOUNT>.1password.eu --email <EMAIL>
  packages:
    - name: op
      version: '~2'
      command: op --version
      help: |
        Please ensure you have the 1Password cli 'op' installed in the required version: %s!

          $ brew update
          $ brew install 1password-cli
```
