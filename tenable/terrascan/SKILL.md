#### Hazards

**Every verb here is read-only** — it scans infrastructure-as-code and reports policy
violations without modifying anything. Nothing needs approval on the grounds of side
effects.

**Its exit status is the finding, so do not discard it.** The binary exits non-zero on
violations and the provider returns that error unchanged, aborting the loop at the
**first** directory with findings — so the directories after it were never scanned.
Treat a failure as a partial result, not a complete report.

**Omitting the path scans every matching directory in the project.** The argument is
optional and repeatable; with none given, the provider walks the working directory for
the mode's marker file and scans each containing directory. That is the intended "scan
everything" case, and on a large repo it is the slow one.

**This command doubles as a linter.** It implements the `Linter` interface
`arbitrary/lint` discovers by type-asserting every registered command, so `lint` picks
it up with no explicit registration and runs all three modes across the project. Its
`fix` parameter is ignored, as a scanner has nothing to fix.

#### Behaviour

Each subcommand is a **directory** scan keyed to a marker file, not a file scan:
`helm` finds `Chart.yaml`, `terraform` finds `main.tf`, `docker` finds `Dockerfile`,
and in each case the *containing directory* is passed as `--iac-dir` with
`--iac-type` set to the mode. So a path argument should name a directory, and
completion offers exactly the directories holding the marker.

Discovery skips `node_modules` and dot-directories, and is cached for the session — a
chart added after the shell started is neither offered nor scanned until the cache is
cleared.

Only `--iac-dir` and `--iac-type` are passed. This tree declares no flags, and neither
additional args nor flags after a `--` separator reach the binary, so upstream options
are unreachable from here; run the binary directly if you need them.

#### Configuration

**This provider has no config key of its own** — there is no `Config` type and no
schema, so there is nothing to set in the project's posh config. What gets scanned is
decided entirely by which marker files exist on disk.

#### Examples

```bash
posh execute {{cmd}} helm path/to/chart
posh execute {{cmd}} terraform
```

#### References

- [terrascan](https://github.com/tenable/terrascan) — the scanner this wraps, and its exit codes
- [`arbitrary/lint`](https://github.com/foomo/posh-providers/blob/main/arbitrary/lint/README.md) — the aggregator that runs this as a linter
