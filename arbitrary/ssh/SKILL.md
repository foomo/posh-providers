#### Hazards

Every verb here manages **long-lived background `ssh` processes** through
[gokazi](https://github.com/foomo/gokazi), not the current shell. `start`
returns as soon as the process is up and the tunnel keeps running after the
command finishes - and, because the process is started with
`context.WithoutCancel`, after the posh shell that started it exits. Nothing
here reports status; use the `gokazi` command to see what is actually running.

**`start` and `stop` with no name act on every configured entry.** The name
argument is optional and repeatable, and when it is omitted the provider
substitutes the full list from the config. So `ssh pfw stop` tears down every
port forward, and `ssh socks5 stop` every tunnel - the plural case is the
default, not an error. Both verbs also iterate and return on the first failure,
so a name that is not in the config aborts the loop and leaves the entries
before it already started or stopped.

`start` is idempotent per name: an entry already running is silently a no-op
(`gokazi.ErrAlreadyRunning` is swallowed), and so is stopping one that is not
running. Neither prints a distinct message, so a second `start` looks the same
as the first.

**Authentication is interactive unless the config avoids it.** `pfw start`
shells out to `ssh -M -n -N -T` with no `BatchMode`, so a key with a passphrase,
an unknown host key or a password prompt blocks the process where an agent
cannot answer it - and stdin is `/dev/null` (`-n`), so it blocks rather than
failing fast. Configure `identityFile`/`identityAgent` and a known host before
running it unattended. `socks5 start` does set `-o BatchMode=yes` along with
`ExitOnForwardFailure` and keepalives, so it fails instead of prompting.

**`port: 0` does not auto-assign for a port forward.** `StartPortForward`
computes a free port into a local variable (`ssh.go:107`) but then builds the
flag from the raw config value (`-L <c.Port>:...`, `ssh.go:117`), so a `0`
reaches `ssh` verbatim and the computed port is only used in the "ready at"
message. The log line therefore names a port nothing is listening on, and the
forward is not usable. Set an explicit `port` for every port forward.
`socks5 start` does use the computed value and auto-assigns correctly.

#### Behaviour

Two further defects in `ssh.go`, neither dangerous but both apt to mislead:

**The socks5 task registration passes a host where a port belongs.** The task
gokazi records for a tunnel is built with `-D <host>` (`ssh.go:85`) instead of
`-D <port>`. What actually runs comes from `StartSocks5Tunnel`, which builds its
own correct command, so this only corrupts the recorded task metadata - but it
is what a `gokazi`-side listing shows, so do not read that listing as the
command that is running.

**A port forward's `hostPort` is the forwarding target, not the SSH port.** Both
structs carry the same three fields, and `socks5` treats `hostPort` as the SSH
server port (`-p`), while `pfw` puts it on the right-hand side of `-L` and has
no way to reach a non-22 SSH port at all. The same config key means two
different things depending on which subtree you are in.

#### Configuration

Config key `ssh` by default, overridable via `WithConfigKey` (and the command
renameable via `CommandWithName`), so confirm both against the project's own
posh config. This provider ships **no `config.schema.json`**, so the key is not
described in the project's `posh.schema.json` - read `config.go`,
`portforward.go` and `socks5tunnel.go`, or the sample in
[`arbitrary/ssh/README.md`](https://github.com/foomo/posh-providers/blob/main/arbitrary/ssh/README.md),
for the field shapes.

Two keys, `portForwards` and `socks5Tunnels`, each a map of name to entry. Those
map keys are the names the `name` argument accepts and completes from; the
entries beneath them are what reach `ssh`. There is no way to define a tunnel
from the command line - the config is the only surface.

Things the config shape cannot tell you: `host` and `username` are passed
through `os.ExpandEnv`, so `$VAR` references in them resolve from the posh
process's environment, and `identityFile` additionally goes through tilde
expansion. An empty `username` or `identityFile` is simply omitted from the
command, falling back to whatever the user's own `~/.ssh/config` says - so the
same posh config behaves differently per machine.

Each configured entry is registered with gokazi under `ssh.pfw.<name>` /
`ssh.socks5.<name>`, which is the id it appears under in the shared process
registry that `gokazi`, `kubeforward`, `dockprox` and `gost` also write to.

#### Examples

```bash
# Start one named port forward; needs an explicit port in the config
posh execute ssh pfw start my-forward

# Start every configured socks5 tunnel
posh execute ssh socks5 start

# Stop one; with no name this stops ALL port forwards
posh execute ssh pfw stop my-forward
```

#### References

- [ssh(1)](https://man.openbsd.org/ssh) - `-L`, `-D`, `-N`, `-M` and `BatchMode`
- [gokazi](https://github.com/foomo/gokazi) - the background process registry these commands write to
- [Provider README](https://github.com/foomo/posh-providers/blob/main/arbitrary/ssh/README.md)
