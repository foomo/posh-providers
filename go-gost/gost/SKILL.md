#### Hazards

Both verbs manage **long-lived background `gost` processes** through
[gokazi](https://github.com/foomo/gokazi), not the current shell. `start` returns
once the process is up and the tunnel keeps running afterwards. Nothing here
reports status; use the `gokazi` command to see what is actually running.

**Omitting the name acts on every configured process.** The argument is optional
on both verbs and leaving it off widens the scope rather than narrowing it, so
`stop` with no name stops every configured process while `stop local` stops
exactly that one. The usage block shows only `[name]...`, which reads as less
work, not more.

What a name actually does is invisible from here. Each maps to a
`gost` invocation with `-C <file>`, and that file decides which ports are bound and
which traffic is proxied where - so a `start` may expose a local port, reach a
production network, or fail on an address already in use, with nothing in the
command name to say which. Read the referenced file before starting a name you
do not recognise.

Each configured entry registers with gokazi under `gost.<name>`, an id in the
shared process registry that `gokazi`, `kubeforward`, `ssh` and `dockprox` also
write to. A `gokazi stop` reaches these processes; nothing here reaches theirs.

#### Behaviour

Both verbs validate each name inside the loop and return on the first error, so
an unknown name aborts with every earlier name already started or stopped - the
validation is not a pre-flight check over the whole list.

The `gost` binary must be on `PATH`; the provider never checks. Processes are
started with the invoking command's context rather than detached with
`context.WithoutCancel`, unlike `arbitrary/ssh`.

#### Configuration

The `gost` key is a **flat map of name to config-file path**, not an object with
fields. Its keys are the names both verbs accept and complete from; each value is
passed as `-C <path>`, resolved relative to the project root.

Field shapes: [`go-gost/gost/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/go-gost/gost/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} start local staging
posh execute {{cmd}} stop local
```

#### References

- [gost documentation](https://gost.run/en/) - the config file format each name points at
- [gokazi](https://github.com/foomo/gokazi) - the background process registry these commands write to
- [Provider README](https://github.com/foomo/posh-providers/blob/main/go-gost/gost/README.md)
