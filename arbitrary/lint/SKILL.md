#### Hazards

**This command's blast radius is invisible from its own tree.** It runs no
linter of its own: it collects every registered posh command that happens to
implement a `Lint` method and invokes it. Which linters those are depends
entirely on what the project's plugin registered, so the same `lint` command
does different work in different projects. `lint` with no argument runs **all**
of them. Read the project's plugin, or the completion list, to find out what
will actually run.

**`lint --fix` is a mutating command.** The flag is forwarded to every linter
that runs, and a linter is free to rewrite files with it - among the providers
here, `terraform-linters/tflint` rewrites `.tf` files in place, while
`tenable/terrascan` and `vsiakka/gherkin-lint` ignore the flag. So the files
`lint --fix` may change are the union of whatever the registered linters choose
to touch, none of it named in this command's usage. Run it on a clean working
tree, and prefer naming a single linter when fixing.

**Every linter runs concurrently with no limit**, all sharing one context. The
first failure cancels the rest, so a failing run leaves an arbitrary subset
completed - and with `--fix`, an arbitrary subset of files already rewritten.
Interleaved output from parallel linters also makes it hard to attribute a
message to its source.

#### Behaviour

Registration is implicit and by type assertion: any command in the project's
`command.Commands` registry with `Name() string` and
`Lint(ctx context.Context, fix bool) error` becomes a linter here. Nothing lists
them in configuration, and a provider's own documentation may not mention that
it doubles as one. Adding an unrelated provider to the registry can therefore
change what `lint` does.

Naming linters filters the set: `lint go tsc` runs exactly those. Matching is
against each linter's `Name()`, which is the command name it registers under.
A name matching no registered linter is warned about individually
(`unknown linter: <name>`) and the linters that did match still run, so a typo
is visible without blocking the rest. If *no* name matches there is nothing to
run and the command fails instead.

A failing linter's exit status propagates, so this command is usable as a CI
gate - subject to the cancellation caveat above.

**The README is out of date.** It shows seven options - `CommandWithGo`,
`CommandWithTSC`, `CommandWithHelm`, `CommandWithESLint`, `CommandWithGherkin`,
`CommandWithTerraform`, `CommandWithTerrascan` - none of which exist. Linters
are not selected by option at all; they are discovered from the registry. The
only option is `CommandWithName`, and `NewCommand` takes `command.Commands`
rather than a cache.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type, no
`config.schema.json` and no `config.base.json`, so there is nothing to set in
the project's posh config and no entry for it in `posh.schema.json`. What runs
is determined by the plugin's command registry; how each linter behaves is
determined by that linter's own configuration file.

The command can be renamed with `CommandWithName`. Note it is constructed with
the command registry itself, so it must be added after the linters it should
discover.

#### Examples

```bash
# Runs every registered linter, concurrently - read-only
posh execute lint

# Only these two
posh execute lint tflint terrascan

# Mutating: each linter that honours --fix will rewrite files
posh execute lint tflint --fix
```

#### References

- [`terraform-linters/tflint`](https://github.com/foomo/posh-providers/blob/main/terraform-linters/tflint/README.md) - honours `--fix`, rewrites `.tf` files
- [`tenable/terrascan`](https://github.com/foomo/posh-providers/blob/main/tenable/terrascan/README.md) and [`vsiakka/gherkin-lint`](https://github.com/foomo/posh-providers/blob/main/vsiakka/gherkin-lint/README.md) - ignore `--fix`
- [Provider README](https://github.com/foomo/posh-providers/blob/main/arbitrary/lint/README.md) - its option list does not exist
