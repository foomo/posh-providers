#### Hazards

**The first argument picks which Grafana server you mutate; nothing in the command name
says which is production.** The env name resolves to a URL, org id and token; everything
after it passes through unchanged. A delete against `prod` and the same line with `dev`
differ by one word, and destructive verbs — deleting dashboards, datasources or folders —
arrive with production credentials and no confirmation from posh. Confirm the env first.

**The tree is not the surface.** Only the env name is modelled; the second argument is a
repeated passthrough, so the entire upstream CLI is reachable and none of it appears in
the usage block. Consult the binary's own `--help`.

**Every invocation resolves a live credential before exec.** Either `token`, expanded
through `envsubst` and — when it's a secret reference — fetched via the 1Password CLI, or
`tokenCmd`, which **executes the configured command line** and takes its stdout as the
token. An ordinary read can trigger a 1Password prompt or run an arbitrary config command.
Neither set fails with `no token provided`. The token stays out of the
process list — but `tokenCmd` runs as a visible child process.

**An env with `kubeforward` set starts a `kubectl port-forward`**, torn down on exit but
left running on a crash between start and defer. Needs `CommandWithKubectl` wired in, else
`kubeforward` is ignored and `GRAFANA_SERVER` falls back to `url`.

#### Behaviour

Flags after a `--` separator are **not** forwarded — dropped rather than reaching the
binary. Put flags before the separator.

#### Configuration

Read under the `gcx` key: a **flat map of environment name to settings**, not an object
with fields; map keys are the names the first argument accepts.

`token` need not be a literal: `$VAR` resolves from posh's environment, a secret-reference
prefix makes it a 1Password lookup. `tokenCmd` is an argv array that gets executed; `token`
wins when both are set.

`kubeforward` (`target`, `namespace`, `port`) routes the server through a port-forward
instead of `url` — env name doubles as kubectl cluster name.

Field shapes: [`grafana/gcx/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/grafana/gcx/config.schema.json).

#### Examples

```bash
# read-only, still resolves the env's credential
posh execute {{cmd}} dev dashboards list

# env is the only thing scoping this - check first
posh execute {{cmd}} prod dashboards delete <uid>
```

#### References

- [Grafana CLI](https://github.com/grafana/gcx) - the binary this wraps; its `--help` is the real surface
- `grafana/gcx/README.md` - plugin wiring only; its config section is stale
