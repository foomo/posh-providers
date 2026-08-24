#### Hazards

**The first argument picks which Grafana server you are mutating, and nothing in the
command name says which is production.** The env name resolves to a URL, an org id and a
token; everything after it passes through unchanged. So a delete against `prod` and the
same line with `dev` differ by one word, and the upstream CLI's destructive verbs —
deleting dashboards, datasources or folders — arrive with production credentials attached
and no confirmation from posh. Confirm the env before any verb that is not a read.

**The tree is not the surface.** Only the env name is modelled; the second argument is a
repeated passthrough, so the entire upstream CLI is reachable and none of it appears in
the usage block. Treat this as the whole Grafana CLI scoped to one server, and consult the
binary's own `--help`.

**Every invocation resolves a live credential before exec.** Either `token`, expanded
through `envsubst` and — when the result is a secret reference — fetched by shelling out
to the 1Password CLI, or `tokenCmd`, which **executes the configured command line** and
takes its stdout as the token. So an ordinary-looking read can trigger a 1Password prompt
or run an arbitrary command from config. An env with neither fails with
`no token provided`. The token is passed as an environment variable, so it stays out of
the process list — but `tokenCmd` runs as a visible child process.

#### Behaviour

Flags after a `--` separator are **not** forwarded — they are silently dropped rather
than reaching the binary. Put flags before the separator.

#### Configuration

Read under the `gcx` key. It is a **flat map of environment name to settings**, not an
object with fields; the map keys are the names the first argument accepts.

`token` need not be a literal: `$VAR` resolves from the posh environment and a
secret-reference prefix makes it a 1Password lookup. `tokenCmd` is an argv array that
gets executed; `token` wins when both are set.

Field shapes: [`grafana/gcx/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/grafana/gcx/config.schema.json).

#### Examples

```bash
# Read-only, but still resolves the env's credential first
posh execute {{cmd}} dev dashboards list

# The env is the only thing scoping this - check it first
posh execute {{cmd}} prod dashboards delete <uid>
```

#### References

- [Grafana CLI](https://github.com/grafana/gcx) - the binary this wraps; its `--help` is the real surface
- `grafana/gcx/README.md` - plugin wiring only; its config section is stale
