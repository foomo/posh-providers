#### Hazards

**This command does not reliably terminate.** The provider hardcodes a
`--out web-dashboard` flag, starting the dashboard server on port **5665**, and the
binary deliberately waits to exit while a browser is open on it. A run that finished its
test can sit there indefinitely, so an agent appears to hang rather than fail. The port
is not configurable through this tree: set `K6_WEB_DASHBOARD_PORT=-1` to disable the
server, or `K6_WEB_DASHBOARD_EXPORT=<file>` for a report at the end. Assume this needs a
human unless one is set.

A load test generates real traffic, and **which system it hits is chosen by the env
argument** — a local run and one hammering production differ by one word. Nothing
validates it: an unknown env silently yields an *empty* variable set rather than an
error, so the scenario runs on its compiled-in defaults. Treat any env resolving to a
shared or production URL as needing explicit approval.

Secret injection writes **decrypted secrets to disk** in the configured path whenever a
matching template exists, and nothing cleans them up — confirm that path is git-ignored.
Without a signed-in 1Password session the injection fails and the run aborts; with no
1Password provider wired in the step is skipped and only pre-existing secret files are
used.

No verb writes project source, state or cloud resources.

#### Behaviour

The scenario argument completes from a walk of the configured path for the project's
naming convention — a script not matching it cannot be completed but still runs if
typed. Secret filenames are derived from the scenario name.

The two secret sources are registered under names that matter: the project-wide file as
`name=all`, the per-scenario one as `name=default`. A scenario reading the default source
gets its own file, not the shared one. Each is passed only when present, so a missing
template silently changes which sources exist rather than erroring.

#### Configuration

Read under the `k6` key. `envs` defines what each environment name points at — read it
before running. `path` determines both which scenarios exist and where decrypted secrets
are written.

Field shapes: [`grafana/k6/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/grafana/k6/config.schema.json).

#### Examples

```bash
# Preferred unattended: no dashboard server, report on disk.
K6_WEB_DASHBOARD_PORT=-1 K6_WEB_DASHBOARD_EXPORT=report.html \
  posh execute {{cmd}} dev checkout

# Extra upstream flags go after the separator.
posh execute {{cmd}} dev checkout -- --vus 10
```

#### References

- [Web dashboard](https://grafana.com/docs/k6/latest/results-output/web-dashboard/) — the wait-for-browser behaviour
- [Secret sources](https://grafana.com/docs/k6/latest/using-k6/secret-source/)
