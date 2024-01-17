# POSH 1Password provider

Integrates 1Password into your shell and adds helpers for your commands.

## Help

```
1Password session helper.

Usage:
  op [command]

Available commands:
  get [id]          Retrieve an entry from your account
  signin            Sign into your 1Password account for the session
  register [email]  Add your 1Password account
```

## Usage

### Plugin

```go
package plugin

type Plugin struct {
  l        log.Logger
  oo       *onepassword.OnePassword
  cacche   cache.Cache
  commands command.Commands
}

func New(l log.Logger) (plugin.Plugin, error) {
  inst := &Plugin{
    l:        l,
    cache:    cache.MemoryCache{},
    commands: command.Commands{},
  }

  // ...

  inst.op, err := onepassword.New(l, inst.cache));
  if err != nil {
    return nil, errors.Wrap(err, "failed to create onepassword")
  }

  // ...

  inst.commands.MustAdd(onepassword.NewCommand(l, onePassword))

  // ...

  return inst, nil
}
```

### Config

Add this to your '.posh.yml' file:

```yaml
onePassword:
  account: <ACCOUNT>
  tokenFilename: .posh/config/.op
```

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
