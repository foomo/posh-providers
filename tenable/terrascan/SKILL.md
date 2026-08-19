#### Hazards

**This command is read-only** - it scans infrastructure-as-code and reports
policy violations without modifying anything. Nothing here needs approval on the
grounds of side effects.

**Its exit status is the finding, so do not discard it.** terrascan exits
non-zero when it finds violations, and the provider returns that error
unchanged, aborting the loop at the **first** directory with findings. So a
failed run means "violations found here", the directories after it were never
scanned, and a clean exit only proves the scanned set was clean up to that
point. Treat a failure as a partial result, not a complete report.

**Omitting the path scans every matching directory in the project.** The
argument is optional and repeatable; with none given, the provider walks the
working directory for the mode's marker file and scans each containing directory
in turn. That is the intended "scan everything" case, but on a large repo it is
also the slow one.

#### Behaviour

Each subcommand is a **directory** scan keyed to a marker file, not a file scan:
`helm` finds `Chart.yaml`, `terraform` finds `main.tf`, `docker` finds
`Dockerfile`, and in each case the *containing directory* is what gets passed to
terrascan as `--iac-dir`, with `--iac-type` set to the mode. So a path argument
should name a directory, and completion offers exactly the directories that
contain the marker.

Discovery skips `node_modules` and any dot-directory, and the result is cached
for the session - a chart or module added after the shell started will not be
offered or scanned until the cache is cleared.

Only `--iac-dir` and `--iac-type` are passed. Nothing else is forwarded: this
tree declares no flags, and neither additional args nor additional flags after a
`--` separator reach the binary. Upstream options like `--policy-type`,
`--severity` or an output format are not reachable from here; run `terrascan`
directly if you need them.

**This command doubles as a linter.** It implements the `Linter` interface that
`arbitrary/lint` discovers, so `lint` picks it up automatically by type-asserting
every registered command - there is no explicit registration. Running `lint`
therefore runs all three modes sequentially across the whole project. Its `fix`
parameter is ignored, as a scanner has nothing to fix.

Note the README's usage snippet does not compile: it calls
`lint.NewCommand(l, cache, lint.CommandWithLinters(cmd))`, but that option does
not exist and `lint.NewCommand` takes a `command.Commands` registry. Adding this
command to that registry is what wires it up.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type, no
`config.schema.json` and no `config.base.json`, so there is nothing to set in
the project's posh config and no entry for it in `posh.schema.json`. What gets
scanned is decided entirely by which marker files exist on disk.

The command can be renamed with `CommandWithName`, and the terrascan binary
invocation swapped with `CommandWithExecTerrascan` - both options on
`NewCommand`.

#### Examples

```bash
# Scan one chart directory
posh execute terrascan helm path/to/chart

# Scan every directory containing a main.tf
posh execute terrascan terraform
```

#### References

- [terrascan](https://github.com/tenable/terrascan) - the scanner this wraps, and its exit codes
- [`arbitrary/lint`](https://github.com/foomo/posh-providers/blob/main/arbitrary/lint/README.md) - the aggregator that runs this as a linter
- [Provider README](https://github.com/foomo/posh-providers/blob/main/tenable/terrascan/README.md) - note its usage snippet does not compile
