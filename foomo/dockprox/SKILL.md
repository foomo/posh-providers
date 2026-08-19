#### Hazards

All three verbs manage a **long-lived background `dockprox` process** through
[gokazi](https://github.com/foomo/gokazi), not the current shell. `start` returns
as soon as the process is up and the proxy keeps running after the command
finishes, holding whatever ports its config binds. Use the `gokazi` command, or
this provider's prompt checker, to see whether it is actually running.

**`stop` cannot stop a process started by `menubar`.** The task registered at
construction carries `Args: [<config path>]`, and gokazi matches a running
process by requiring every registered argument to appear in its command line
(`findProcess`, arg gate). `start` runs
`dockprox serve --config <config path>`, which contains it and is therefore
stoppable; `menubar` runs a bare `dockprox menubar`, which does not - so after
`menubar`, `stop` reports nothing running and the process is left behind for the
user to kill by hand. Verified against the matching rule, not by reading alone.

The same mismatch means `start` and `menubar` are **not** mutually exclusive
despite sharing one gokazi id. gokazi refuses a second `start` only when it can
*find* the first, so `menubar` followed by `start` launches a second dockprox
rather than reporting `already running` - two proxies competing for the same
ports.

`menubar` starts a **desktop menubar application**, not a headless proxy. It is
not useful under an agent and, per the above, not stoppable through this command
either; leave it to a human at a graphical session.

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

Both `start` and `menubar` register under the single gokazi id `dockprox`, so
unlike the sibling background providers there is no per-name selection: this
command manages exactly one process. It shares the registry with `gokazi`,
`kubeforward`, `ssh` and `gost`, so a bare `gokazi stop` reaches this process
too.

`Checker()` builds its **own** `gokazi.Gokazi` and registers a task named
`dockprox` with no `Args`, which is a looser match than the command's own task -
so the prompt can report `Running` for a menubar process that `stop` will not
find. A green prompt entry is not a promise that `stop` will work.

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
