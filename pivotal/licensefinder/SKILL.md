#### Hazards

Nothing here touches infrastructure or credentials, and the two reporting verbs —
a bare `licensefinder` (which runs `action_items`) and `report` — are read-only and
safe to run unattended for orientation.

What the `add` and `remove` subtrees do write is the **decisions file** named by
config, in place, on every invocation. That file is the project's record of which
licenses have been approved and which dependencies are exempt, so it is normally
version-controlled and reviewed. `add permitted <license>` approves a license for
**every** dependency that uses it, now and in future, and `add ignored <dependency>`
removes a dependency from license checking altogether — both widen what the project
accepts without any diff being shown. `remove` deletes such a decision, which can
make a previously-passing check start failing.

So these are policy changes rather than build steps: they should be proposed to a
human and committed deliberately, not applied to clear a failing check. The `--who`
and `--why` flags exist precisely because the decision is meant to be attributable;
record both when a human has asked for the change.

A bare `licensefinder` exits **1** when unapproved dependencies exist — that is the
intended signal, not a tool failure, so do not "fix" it by adding decisions. Note the
inverse trap: when nothing is recognised at all it prints `No dependencies
recognized!` and exits **0**, so a green result can mean "scanned nothing" rather than
"everything approved". Given the aggregation defect below, check the logged path list
rather than trusting the exit code.

`Validate` requires the `license_finder` binary on PATH and, unlike most providers
here, points at `brew install licensefinder` rather than an ownbrew package — the
provider README declares no ownbrew entry, so the tool has to be installed on the host
before any of these verbs work.

#### Behaviour

**`action_items` and `report` cover every discovered source directory, not the
current one.** The provider walks for each configured `sources` manifest and passes
every match after `--aggregate_paths`, so a multi-project repository yields one
combined report. The path list is logged before the run; to confirm what upstream
would scan, compare it against `license_finder project_roots`, which is not in this
tree and has to be run directly.

Discovery walks the checkout from `.` for each configured `sources` filename,
ignoring dotted directories, `vendor` and `node_modules`, and uses each match's
**containing directory**. So `sources` names manifest files while the tool receives
directories, and a manifest inside `vendor/` or `node_modules/` is invisible by
design. The result is cached per shell session and per filename, so a newly added
manifest needs a `cache clear`. The aggregated list is logged before each run, which
is the quickest way to see what was actually included.

`--who` and `--why` are declared only on the two `add` leaves, but upstream marks its
`remove` verbs `auditable` too — so a removal *can* be attributed and this tree gives
no way to do it. Pass them after `--` if a removal needs a recorded reason. Only
flags actually typed are forwarded (`fs.Visited().Args()`), so nothing posh-internal
leaks to the binary.

Both `add` and `remove` are variadic upstream (`def add(*licenses)`), and the
provider forwards every argument after the verb, so several names can be given in one
invocation even though the tree declares a single `name` argument.

`--log-directory` and `--decisions-file` are appended by the provider on **every**
invocation from config, so those two cannot be overridden by typing them — a
duplicate would be passed after the provider's own. Anything else goes after `--`
and reaches the binary.

The verb names in this tree are not upstream's: `add permitted` maps to
`permitted_licenses add`, `list ignored` to `ignored_dependencies list`, and so on.
A bare `licensefinder` maps to `action_items`, upstream's own default.

#### Configuration

Read under the **`licenseFinder`** key by default — note the camel case, which
differs from the command name; `CommandWithConfigKey` renames it and
`CommandWithName` renames the command, both `CommandOption`s on `NewCommand`.
Confirm against the project's own config file. Field shapes:
[`pivotal/licensefinder/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/pivotal/licensefinder/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`sources` decides the entire scope of every scan, and `decisionsPath` is the file the
mutating verbs rewrite — read both before running anything under `add` or `remove`.

#### Examples

```bash
# Orientation, read-only. Non-zero exit means unapproved dependencies exist.
posh execute licensefinder
posh execute licensefinder report

# Inspect the current policy before changing it.
posh execute licensefinder list permitted
posh execute licensefinder list ignored

# Policy changes — attribute them, and only when a human asked.
posh execute licensefinder add permitted MIT --who "Jane Doe" --why "OSI approved"
posh execute licensefinder add ignored some-internal-pkg --who "Jane Doe" --why "First-party"
```

#### References

- [LicenseFinder README](https://github.com/pivotal/LicenseFinder#readme)
- [Approving dependencies](https://github.com/pivotal/LicenseFinder#approving-dependencies) — what `add permitted` / `add ignored` mean
- [Decisions file](https://github.com/pivotal/LicenseFinder#decisions-file) — the format of the file these verbs rewrite
- [Provider README](https://github.com/foomo/posh-providers/blob/main/pivotal/licensefinder/README.md)
