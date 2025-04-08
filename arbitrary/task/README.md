# POSH task provider

> Run simple tasks definitions.

## Usage

### Plugin

```go
package main

type Plugin struct {
  l        log.Logger
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  var err error
  inst := &Plugin{
    l:        l,
    commands: command.Commands{},
  }

	// ...

  inst.commands.Add(task.NewCommand(l))

  // ...

  return inst, nil
}
```

### Config

```yaml
## Open
tasks:
  init:
    cmds: ['posh execute bun install', 'posh execute go mod tidy']
  brew-nss:
    hidden: true
    confirm: If you're using Firefox, do you want me to install 'nss'?
    cmds: ['brew install nss']
  mkcert-install:
    hidden: true
    confirm: Do you need me to install the mkcert root certificate (only required once)?
    cmds: ['posh execute mkcert install']
  k3d-up:
    deps: ['brew-nss', 'mkcert-install']
    cmds:
      - posh execute mkcert generate
      - posh execute k3d up local
      - posh execute cache clear
```

