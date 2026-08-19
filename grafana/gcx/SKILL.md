#### Hazards

**The first argument picks which Grafana server you are mutating, and nothing
in the command name says which is production.** The env name resolves to a URL,
an org id and a token from the config; everything after it goes to `gcx`
unchanged. So `gcx prod dashboards delete <uid>` and the same line with `dev`
differ by one word, and gcx's own destructive verbs - deleting dashboards,
datasources or folders - arrive with production credentials already attached and
no confirmation step from posh. Confirm the env before running any verb that is
not a read.

**The tree is not the surface.** Only the env name is modelled here; the second
argument is a repeated passthrough, so the entire `gcx` CLI is reachable and
none of it appears in the usage block. Treat this as "the whole Grafana CLI,
scoped to one server" rather than as the two arguments shown. Consult gcx's own
`--help` for what is actually available.

**Running any subcommand resolves a live credential.** Every invocation reads
the env's token before exec: either `token`, expanded through `envsubst` and -
when it starts with `op://` - fetched by shelling out to `op get`, or
`tokenCmd`, which **executes the configured command line** and takes its stdout
as the token. Both mean an ordinary-looking `gcx <env> dashboards list` can
trigger a 1Password prompt or run an arbitrary command from the config. An env
with neither field fails with `no token provided` before gcx runs.

The resolved token is passed to gcx as the `GRAFANA_TOKEN` environment variable,
not on the command line, so it does not appear in the process list - but
`tokenCmd` itself does run as a visible child process.

#### Behaviour

**The README is out of date; do not follow it.** It documents a `path` config
key, one gcx config file per environment, and a `GCX_CONFIG=<file>` invocation.
None of that exists in the code: the config is a map of env name to settings,
and the provider passes `GRAFANA_SERVER`, `GRAFANA_ORG_ID` and `GRAFANA_TOKEN`
as environment variables instead. Read `config.schema.json` or `configenv.go`,
not the README's config section.

Four environment variables are set on every call beyond the credentials, and
they are not configurable: `DO_NOT_TRACK=1`, `GCX_TELEMETRY=0` and
`GCX_NO_UPDATE_NOTIFIER=1` suppress telemetry and update checks, so gcx behaves
more quietly here than when run by hand. An env's own `env` list is applied
after these, so it can override any of them.

Completion of the second argument shells out to the real `gcx __complete`
binary, so tab completion needs `gcx` installed and may make it do work. It does
not depend on the selected environment - the command tree is the same for all of
them.

Additional flags after a `--` separator are **not** forwarded. `Execute` passes
`r.Args()`, `r.Flags()` and `r.AdditionalArgs()` but never `AdditionalFlags()`,
so a flag placed after `--` is silently dropped rather than reaching gcx. Put
flags before the separator.

#### Configuration

Config key `gcx` by default, overridable via `CommandWithConfigKey` (and the
command renameable via `CommandWithName`), so confirm both against the project's
own posh config.

The key is a **flat map of environment name to settings** - not an object with
fields. The map keys are the names the first argument accepts and completes
from.

Field shapes: [`grafana/gcx/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/grafana/gcx/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

Two things the schema cannot say: `token` is not necessarily a literal - it goes
through `envsubst`, so `$VAR` references resolve from the posh process's
environment, and an `op://` prefix turns it into a 1Password lookup. `tokenCmd`
is an argv array that gets executed, with its trimmed stdout used as the token;
`token` wins when both are set.

#### Examples

```bash
# Read-only, but still resolves the env's credential first
posh execute gcx dev dashboards list

# The env is the only thing scoping this - check it before any destructive verb
posh execute gcx prod dashboards delete <uid>
```

#### References

- [gcx](https://github.com/grafana/gcx) - the Grafana CLI this wraps; its `--help` is the real command surface
- [Provider README](https://github.com/foomo/posh-providers/blob/main/grafana/gcx/README.md) - plugin wiring only; its config section is stale
