#### Hazards

All three verbs manage a **long-lived background `dockprox` process** through
[gokazi](https://github.com/foomo/gokazi), not the current shell. `start` returns
as soon as the process is up and the proxy keeps running after the command
finishes, holding whatever ports its config binds. Use the `gokazi` command, or
this provider's prompt checker, to see whether it is actually running.

`menubar` starts a **desktop menubar application**, not a headless proxy. It
needs a graphical session, so it is not useful under an agent; leave it to a
human. `stop` does stop it.

What the proxy actually does is invisible from here. `start` passes the
configured file to `dockprox serve`, and that file decides which ports are bound
and which containers traffic is routed to - so a `start` may expose local ports
or fail on an address already in use, with nothing in the command name to say
which. Read the referenced file before starting it in a project you do not know.

`dockprox` itself must be on `PATH`; the provider never checks, and a missing
binary surfaces as a start failure after the one-second liveness wait.

#### Behaviour

There are no arguments, flags or names anywhere in this tree - the whole
invocation comes from one config key, and unmatched arguments are not forwarded
to the `dockprox` binary. Whatever you type after the verb is ignored, so there
is no way to reach an upstream flag from here.

`start` and `menubar` register separate gokazi ids (`dockprox.serve` and
`dockprox.menubar`) because gokazi identifies a process by its arguments and the
two run different command lines. `stop` tries both, so it stops whichever is
running, and each verb refuses to start when its own process is already up. There
is still no per-name selection: this command manages one proxy. The registry is
shared with `gokazi`, `kubeforward`, `ssh` and `gost`, so a bare `gokazi stop`
reaches these processes too.

`Checker()` builds its **own** `gokazi.Gokazi` and registers a task named
`dockprox` with no `Args`, a looser match that reports either process as
`Running`.

#### Configuration

Config key `dockprox` by default, overridable via `WithConfigKey` (and the
command renameable via `CommandWithName`), so confirm both against the project's
own posh config.

The key holds a single field, `config`: the path to the dockprox config file,
passed to `dockprox serve --config`. It is used verbatim, with no environment
expansion and no existence check, and it is also what makes a `start`ed process
findable - so changing it while a proxy is running orphans that process.

Field shapes: [`foomo/dockprox/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/dockprox/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

#### Examples

```bash
# Start the proxy in the background from the configured config file
posh execute dockprox start

# Stop it again - only works for a process started by `start`
posh execute dockprox stop
```

#### References

- [dockprox](https://github.com/foomo/dockprox) - the proxy and its config file format
- [gokazi](https://github.com/foomo/gokazi) - the background process registry this command writes to
- [Provider README](https://github.com/foomo/posh-providers/blob/main/foomo/dockprox/README.md)
