#### Hazards

**This command does not reliably terminate.** The provider hardcodes
`--out web-dashboard`, which starts k6's built-in dashboard web server on port
**5665**, and k6 deliberately waits to exit for as long as a browser window is open on
it. So a run that finishes its test can still sit there indefinitely, and an agent
invoking this will appear to hang rather than fail. The port is not configurable
through this tree; set `K6_WEB_DASHBOARD_PORT=-1` in the environment to disable the
server (the extension treats any negative port as off), or
`K6_WEB_DASHBOARD_EXPORT=<file.html>` to get a report written at the end instead.
Assume this verb needs a human unless one of those is set.

The point of a load test is to generate traffic against a real system, and **which
system is chosen by the `<env>` argument** — so a local invocation and one that
hammers production differ by one word. Nothing validates that argument: an unknown
env silently yields an *empty* variable set rather than an error, so the scenario runs
with whatever defaults it has compiled in, which may not be the target you intended.
Confirm the env against the project's config before running, and treat any env
resolving to a shared or production URL as needing explicit approval.

`op inject` writes **decrypted secrets to disk** inside the configured k6 path, at
`all.k6.secret` and `<scenario>.secret`, whenever the matching `.tpl` exists. Those
files are left behind after the run — nothing cleans them up — and they are plaintext.
Confirm the k6 path is git-ignored, and require a signed-in 1Password CLI session:
without one, `op inject` fails and the run aborts. Passing no `*onepassword.OnePassword`
skips the whole step, in which case only pre-existing `.secret` files are used.

Everything else here is read-only with respect to the project: no verb writes source,
state or cloud resources. The root has no subcommands, so this is the whole surface —
but it is not a passthrough either; unmatched arguments become the two positional
arguments rather than being forwarded, and extra upstream flags need `--`.

#### Behaviour

`--no-usage-report` and `--out web-dashboard` are appended by the provider on **every**
run and appear nowhere in the usage block. The dashboard's other options are only
reachable through `K6_WEB_DASHBOARD_*` environment variables — `PORT`, `HOST`, `PERIOD`,
`OPEN` (default false, so no browser is launched for you), `EXPORT` and `RECORD`.

`<scenario>` completes from a walk of the configured path for **`*.k6.js` files**,
with the path prefix stripped, so completion reflects that naming convention rather
than anything k6 knows about. A script not matching `*.k6.js` cannot be completed,
though it still runs if typed — the value is joined onto the configured path verbatim.
Secret filenames are derived by trimming `.js`, so `checkout.k6.js` pairs with
`checkout.k6.secret`, matching the `all.k6.secret` convention.

Two secret sources are wired when the files exist, and the names matter: the
project-wide one is registered as `name=all` and the per-scenario one as
`name=default`. A scenario reading the default secret source therefore gets its own
file, not the shared one, and each is only passed when present — so a missing `.tpl`
or a skipped `op inject` silently changes which sources k6 has, rather than erroring.

Env values are upper-cased before being exported (`url:` becomes `URL=`), so config
keys are case-insensitive in practice and the scenario sees them via `__ENV`. They are
set on the child process rather than interpolated into the command line, so a value
containing spaces or shell metacharacters survives intact — unlike the scenario path
and the secret-source flags, which are arguments and do go through `sh -c`.

Only flags actually typed are forwarded (`fs.Visited().Args()`), so nothing
posh-internal leaks to the binary; `--verbose` is a genuine k6 flag. There is no
`Validate`, so a missing `k6` binary surfaces as a shell error rather than a helpful
message — unlike most sibling providers.

#### Configuration

Read under the `k6` key by default; `CommandWithConfigKey` renames it and
`CommandWithName` renames the command, both `CommandOption`s on `NewCommand`. Confirm
against the project's own config file. Field shapes:
[`grafana/k6/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/grafana/k6/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`envs` is where the target system for each environment name is defined — read it to
learn what a given `<env>` actually points at before running a load test. `path`
determines both which scenarios exist and where decrypted secrets are written.

#### Examples

```bash
# Run a scenario against the dev environment. Note this may not exit on its own;
# see Hazards.
posh execute k6 dev checkout.k6.js

# Preferred for unattended use: no dashboard server, HTML report on disk.
K6_WEB_DASHBOARD_PORT=-1 K6_WEB_DASHBOARD_EXPORT=k6-report.html \
  posh execute k6 dev checkout.k6.js

# Extra k6 flags go after the separator.
posh execute k6 dev checkout.k6.js -- --vus 10 --duration 30s
```

#### References

- [Using k6](https://grafana.com/docs/k6/latest/using-k6/)
- [k6 environment variables](https://grafana.com/docs/k6/latest/using-k6/environment-variables/) — how `envs` values reach the scenario as `__ENV`
- [k6 web dashboard](https://grafana.com/docs/k6/latest/results-output/web-dashboard/) — the `K6_WEB_DASHBOARD_*` options and the wait-for-browser behaviour
- [k6 secret sources](https://grafana.com/docs/k6/latest/using-k6/secret-source/)
- [`op inject`](https://developer.1password.com/docs/cli/reference/commands/inject)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/grafana/k6/README.md)
