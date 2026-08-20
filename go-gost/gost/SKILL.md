#### Hazards

Both verbs manage **long-lived background `gost` processes** through
[gokazi](https://github.com/foomo/gokazi), not the current shell. `start`
returns once the process is up and the tunnel keeps running after the command
finishes. Nothing here reports status; use the `gokazi` command to see what is
actually running.

**Omitting the name acts on every configured process.** The `name` argument is
optional on both verbs, and leaving it off widens the scope rather than
narrowing it - the usage block shows only `[name]...`, which reads as less work,
not more:

- `gost stop` - stops every configured process
- `gost stop local` - stops exactly `local`
- `gost stop local staging` - stops exactly those two

The same holds for `start`.

Both verbs iterate and return on the first error, so an unknown name aborts the
loop with everything before it already started or stopped. `start` validates
each name against the config first, so a typo fails before anything runs - but
only for the names it actually reached.

`gost` itself must be on `PATH`; the provider never checks. The process is
started with the context of the invoking command, so unlike `arbitrary/ssh`
these tunnels are tied to that context rather than detached with
`context.WithoutCancel`.

What a config file actually does is invisible from here. Each name maps to a
`gost -C <file>` invocation, and that file decides which ports are bound and
which traffic is proxied where - so a `start` may expose a local port, reach a
production network, or fail on an address already in use, with nothing in the
command name to tell you which. Read the referenced file before starting a name
you do not recognise.

#### Configuration

Config key `gost` by default, overridable via `WithConfigKey` (and the command
renameable via `CommandWithName`), so confirm both against the project's own
posh config.

The key is a **flat map of name to config-file path** - not an object with
fields. The map keys are the names `start` and `stop` accept and complete from;
each value is passed to gost as `-C <path>` and resolved relative to the project
root the posh shell runs in.

Field shapes: [`go-gost/gost/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/go-gost/gost/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

Each configured entry is registered with gokazi under `gost.<name>`, which is
the id it appears under in the shared process registry that `gokazi`,
`kubeforward`, `ssh` and `dockprox` also write to.

#### Examples

```bash
# Start every configured gost process
posh execute gost start

# Named processes are acted on exactly
posh execute gost start local staging

# Stop one named process
posh execute gost stop local
```

#### References

- [gost documentation](https://gost.run/en/) - the config file format each name points at
- [gokazi](https://github.com/foomo/gokazi) - the background process registry these commands write to
- [Provider README](https://github.com/foomo/posh-providers/blob/main/go-gost/gost/README.md)
