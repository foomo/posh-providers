#### Hazards

Every verb manages **long-lived background `ssh` processes** through
[gokazi](https://github.com/foomo/gokazi), not the current shell. `start` returns
as soon as the process is up, and since it is started with
`context.WithoutCancel` the tunnel outlives the posh shell. Nothing reports
status; use the `gokazi` command to see what is running.

**`start` and `stop` with no name act on every configured entry** - the name is
optional, and when omitted the full config list is substituted, so a bare `stop`
tears down every forward or tunnel. Both verbs return on the first failure,
leaving earlier entries already started or stopped.

**`pfw start` can block on an interactive prompt.** It runs the `ssh` binary with
`-M -n -N -T` and no `BatchMode`, so a passphrase, unknown host key or password
prompt hangs the process - and with stdin at `/dev/null` (`-n`) it blocks rather than failing
fast. Configure `identityFile`/`identityAgent` and a known host before running it
unattended. `socks5 start` sets `BatchMode=yes` and `ExitOnForwardFailure`, so it
fails instead.

**`port: 0` auto-assigns a free local port**, in both subtrees, reported only in
the "ready at" log line - read it to learn where to connect.

#### Behaviour

`start` is idempotent per name: an already-running entry is a silent no-op, as is
stopping one not running, with no distinct message either way.

`hostPort` means two different things per subtree. For `socks5` it is the SSH
server port (`-p`); for `pfw` it is the right-hand side of `-L`, and `host` is
both the SSH destination and the forward target - so `pfw` only forwards to a
port on the bastion itself and cannot reach a non-22 SSH port.

An empty `username` or `identityFile` is omitted entirely, falling back to the
user's own `~/.ssh/config` - so one posh config behaves differently per machine.
`host` and `username` also expand `$VAR`.

#### Configuration

The `ssh` key holds `portForwards` and `socks5Tunnels`, each a map of name to
entry; those keys are what the `name` argument accepts. This provider ships **no
`config.schema.json`**, so the key is absent from the project's
`posh.schema.json` - read `config.go`, `portforward.go` and `socks5tunnel.go` for
field shapes.

Entries register with gokazi as `ssh.pfw.<name>` / `ssh.socks5.<name>` in the
shared registry `kubeforward`, `dockprox` and `gost` also use.

#### Examples

```bash
posh execute {{cmd}} pfw start my-forward
posh execute {{cmd}} socks5 start
posh execute {{cmd}} pfw stop my-forward
```

#### References

- [`ssh(1)`](https://man.openbsd.org/ssh) - `-L`, `-D`, `BatchMode`
- [gokazi](https://github.com/foomo/gokazi)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/arbitrary/ssh/README.md)
