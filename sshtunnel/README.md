SSH Tunnel Manager
==================

A CLI tool to manage multiple SSH tunnels via configuration, supporting password-based authentication, SSH-key authentication, 1Password integration, and optional sudo access.

Features
--------

*   Define multiple SSH tunnels in .posh.yaml configuration file.

*   Support for:

    *   1Password CLI for secure password retrieval

    *   Plaintext passwords or Paths

    *   SSH private key authentication

*   Automatic check if tunnels are running.

*   Automatic check if local port not in use.

*   Automatic check if target proxy port is open.

*   Start and stop tunnels with ease.

*   Optional sudo escalation per tunnel.


Config
-------------

Define your tunnels in .posh.yaml file:

```yaml
sshTunnel:
  socketsPath: devops/config/ssh
  tempDir: "devops/tmp" # used to store tmp keys if privateKey is a op vault
  tunnels:
    - name: "cluster-prod-tunnel"
      localPort: 10443
      targetProxyHost: "192.168.181.96"
      targetProxyPort: 8001
      targetUsername: "administrator"
      targetHost: "78.85.45.111"
      targetAuth:
        type: sshpass
        password: <% op "account" "vault" "item" "field" %> # or plaintext
    - name: "harbor-prod-tunnel"
      sudo: true
      localPort: 443
      targetProxyHost: "core.harbor.domain"
      targetProxyPort: 443
      targetUsername: "administrator"
      targetHost: "98.85.56.25"
      targetAuth:
        type: key
        privateKey: "path/to/your/private_key" # or op vault
```


### Field Description

| Field                   | Type     | Description                                                                       |
| ----------------------- | -------- | --------------------------------------------------------------------------------- |
| `socketsDir`            | string   | Directory where SSH control socket and config files will be stored.               |
| `tempDir`               | string   | Directory where temp keys/config files will be stored.                                 |
| `tunnels`               | list     | List of tunnel configurations.                                                    |
| `name`                  | string   | Unique tunnel name.                                                               |
| `sudo`                  | bool     | Whether to run SSH commands with `sudo`.                                          |
| `localPort`             | int      | Local port to bind the tunnel.                                                    |
| `targetProxyHost`       | string   | Hostname or IP of the internal target service.                                    |
| `targetProxyPort`       | int      | Port of the internal target service.                                              |
| `targetUsername`        | string   | SSH username for the remote host.                                                 |
| `targetHost`            | string   | SSH target host (hostname or IP).                                                 |
| `targetAuth`            | object   | Authentication method configuration.                                              |
| `targetAuth.type`       | enum     | `"sshpass"`, `"key"`.                                                             |
| `targetAuth.password`   | string   | Password string or 1Password CLI template (required for `"sshpass"`).             |
| `targetAuth.privateKey` | string   | Path to SSH private key or 1Password CLI template (required for `"key"`).         |

Authentication
--------------

*   **1Password CLI (op)**: securely fetch passwords from your vault.

*   **Plaintext**: specify the password or private key path directly in the YAML (not recommended for production).


Example 1Password YAML entry:

```yaml
targetAuth:
  type: sshpass
  password: <% op "account" "vault" "item" "field" %>
```
The `targetAuth` is optional and if it's not defined in YAML file the prompt will ask you for the credential

Installation
---------
Define a new command in `".posh/internal/plugin.go"`

```go
type Plugin struct {
	l         log.Logger
	cache     cache.Cache
	op        *onepassword.OnePassword
	sshTunnel *sshtunnel.SSHTunnel
	commands  command.Commands
}
```

Then in the constructor
```go
func New(l log.Logger) (plugin.Plugin, error) {
	inst := &Plugin{
		l:        l,
		cache:    &cache.MemoryCache{},
		commands: command.Commands{},
	}

	if value, err := onepassword.New(l, inst.cache); err != nil {
		return nil, errors.Wrap(err, "failed to create onepassword")
	} else {
		inst.op = value
	}

	if value, err := sshtunnel.New(l); err != nil {
		return nil, errors.Wrap(err, "failed to create sshTunnel")
	} else {
		inst.sshTunnel = value
	}

	// add commands
	inst.commands.MustAdd(onepassword.NewCommand(l, inst.op))
	inst.commands.MustAdd(sshtunnel.NewCommand(l, inst.sshTunnel, inst.cache, inst.op))

	return inst, nil
}
```
Now you need to rebuild the shell with `make shell.rebuild`

CLI Usage
---------
```bash
sshtunnel <tunnel-name> <cmd>
```

### Cmds

| Command | Description                      |
| ------- | -------------------------------- |
| `start` | Starts the specified SSH tunnel. |
| `stop`  | Stops the specified SSH tunnel.  |


Notes
-----

*   SSH tunnels use **control sockets** for master connections.

*   Any missing or inactive tunnel is automatically reported as “not running”.

*   Running SSH commands are quiet by default unless debugging is enabled.
