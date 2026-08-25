#### Hazards

Both verbs manage a **long-lived background `dockprox` process** through
gokazi, not the current shell. `start` returns as soon as the process is up and
the proxy keeps running afterwards, holding whatever ports its config binds. Use
the `gokazi` command, or this provider's prompt checker, to see whether it is
running.

What the proxy does is invisible from here. `start` passes the configured file
to the `dockprox` binary's `serve` verb, and that file decides which ports are
bound and where traffic is routed - so a `start` may expose local ports or fail
on an address already in use, with nothing in the command name to say which.
Read that file before starting it in a project you do not know.

The registry is shared with `gokazi`, `kubeforward`, `ssh` and `gost`, so a bare
`gokazi stop` reaches these processes too.

The `dockprox` binary must be on `PATH`; a missing one surfaces as a start
failure after gokazi's one-second wait for the process to appear.

#### Behaviour

There are no arguments, flags or names in this tree - the whole invocation comes
from one config key, and unmatched arguments are not forwarded. Whatever you
type after the verb is ignored, so no upstream flag is reachable from here.

`start` refuses to start when the proxy is already up; `stop` on a proxy that is
not running is a no-op, not an error. This command manages one proxy.

`Checker()` builds its **own** registry instance with a looser match on the
`dockprox` process name.

#### Configuration

The `dockprox` key of this project's posh config holds a single field, `config`:
the path to the proxy's own config file, forwarded as `--config`. It is used
verbatim, with no environment expansion and no existence check, and it is also
part of what makes a started process findable again - so changing it while a
proxy is running orphans that process.

Field shapes: [`foomo/dockprox/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/dockprox/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} start
posh execute {{cmd}} stop
```

#### References

- https://github.com/foomo/dockprox for the proxy and its config file format
- `foomo/dockprox/README.md` for plugin wiring
