# POSH doctl provider

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

  inst.Commands().MustAdd(mkcert.NewCommand(l))

  // ...

  return inst, nil
}
```

### Config

```yaml
## mkcert
mkcert:
  certificatePath: .posh/config/certs
  certificates:
    - name: foomo.org
      names:
        - foomo.org
        - *.foomo.org
        - localhost
        - 127.0.0.1
        - ::1
```

### Ownbrew

To install binary locally, add:

```yaml
ownbrew:
  packages:
    ## https://github.com/FiloSottile/mkcert/releases
    - name: mkcert
      tap: foomo/filosottile/mkcert
      version: 1.4.4
```
