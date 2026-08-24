#### Hazards

Nothing here touches infrastructure or credentials, and the two reporting verbs — a bare
`{{cmd}}` and `report` — are read-only and safe to run unattended for orientation.

The add and remove subtrees rewrite the **decisions file** named by config, in place. That
file is the project's record of which licenses are approved and which dependencies are
exempt, so it is normally version-controlled and reviewed. Permitting a license approves
it for **every** dependency that uses it, now and in future; ignoring a dependency drops
it from checking altogether. Both widen what the project accepts without showing a diff,
and removing a decision can make a passing check start failing.

These are policy changes, not build steps: propose them to a human rather than applying
them to clear a failing check. The `--who` and `--why` flags exist because the decision is
meant to be attributable.

A bare `{{cmd}}` exits **1** when unapproved dependencies exist — the intended signal,
not a tool failure, so do not "fix" it by adding decisions. The inverse trap: when
nothing is recognised it prints `No dependencies recognized!` and exits **0**, so a green
result can mean "scanned nothing" rather than "everything approved". Check the logged
path list, not just the exit code.

#### Behaviour

**The reporting verbs cover every discovered source directory, not the current one.** The
provider walks the checkout for each configured `sources` manifest and aggregates every
match, so a multi-project repository yields one combined report. The path list is logged
before each run — the quickest way to see what was included.

Discovery ignores dotted directories, `vendor` and `node_modules`, and uses each match's
**containing directory**. So `sources` names manifest files while the tool receives
directories, and a manifest inside `vendor/` is invisible by design. Results are cached
per session, so a newly added manifest needs a `cache clear`.

#### Configuration

Read under the **`licenseFinder`** key — camel case, unlike the command name. `sources`
decides the scope of every scan and `decisionsPath` is the file the mutating verbs
rewrite; read both before changing policy.

Field shapes: [`pivotal/licensefinder/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/pivotal/licensefinder/config.schema.json).

#### Examples

```bash
# Orientation, read-only. Non-zero exit means unapproved dependencies exist.
posh execute {{cmd}} report

# Policy change - attribute it, and only when a human asked.
posh execute {{cmd}} add permitted MIT --who "Jane Doe" --why "OSI approved"
```

#### References

- [Approving dependencies](https://github.com/pivotal/LicenseFinder#approving-dependencies)
- [Decisions file](https://github.com/pivotal/LicenseFinder#decisions-file) — the format these verbs rewrite
