#### Hazards

**`--fix` rewrites your terraform files in place.** Without it this command is a
read-only linter; with it, tflint edits the `.tf` files in every directory it
processes. There is no diff, no confirmation and no backup, so run it on a clean
working tree where `git diff` can show what changed. Combined with an omitted
path - see below - a single `tflint --fix` can rewrite terraform across the
whole project.

**Omitting the path lints every terraform module in the project.** The argument
is optional and repeatable; with none given, the provider walks the working
directory for every `main.tf` and processes each containing directory. That is
the intended "lint everything" case, and it is also what `--fix` applies to.

**Exit status is the finding, and the loop stops at the first one.** tflint
exits non-zero when it reports issues, and the provider returns that error
immediately, so the directories after it are never linted. A failure means
"issues found here, the rest is unchecked" - a partial result, not a complete
report. With `--fix`, it also means earlier directories may already have been
rewritten.

**This command is also an `arbitrary/lint` linter, and `lint` can pass `fix`
through.** It implements the interface `arbitrary/lint` discovers by type
assertion, so it runs as part of any `lint` invocation with no explicit
registration - and `lint --fix` forwards the flag, meaning that command can
rewrite terraform files too. Unlike `tenable/terrascan`, which ignores the
parameter, this provider honours it.

#### Behaviour

Discovery is by marker file: directories containing a `main.tf`. The *directory*
is what tflint runs in - the provider sets the working directory rather than
passing a path argument - so a path given here should name a directory, and
completion offers exactly those that contain a `main.tf`. A terraform module
without a `main.tf` is invisible to both completion and the default sweep.

The discovered list skips `node_modules` and any dot-directory, and is cached
for the session, so a module added after the shell started is not offered or
linted until the cache is cleared.

Only `--fix` is forwarded, and only when set. This tree declares no other flags,
and neither additional args nor additional flags after a `--` separator reach
the binary - so upstream options like `--format`, `--minimum-failure-severity`
or `--config` are not reachable from here. tflint still reads a `.tflint.hcl`
from the directory it runs in, which is the way to configure rules.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type, no
`config.schema.json` and no `config.base.json`, so there is nothing to set in
the project's posh config and no entry for it in `posh.schema.json`. What gets
linted is decided by which directories contain a `main.tf`; how it is linted is
decided by tflint's own `.tflint.hcl`.

The command can be renamed with `CommandWithName`, and the tflint binary
invocation swapped with `CommandWithExecTflint` - both options on `NewCommand`.

#### Examples

```bash
# Read-only, one module
posh execute tflint path/to/module

# Read-only sweep of every module with a main.tf
posh execute tflint

# Rewrites .tf files in place, across every module - needs approval
posh execute tflint --fix
```

#### References

- [tflint](https://github.com/terraform-linters/tflint) - the linter this wraps, its exit codes and `.tflint.hcl`
- [`arbitrary/lint`](https://github.com/foomo/posh-providers/blob/main/arbitrary/lint/README.md) - the aggregator that runs this, and forwards `--fix`
- [Provider README](https://github.com/foomo/posh-providers/blob/main/terraform-linters/tflint/README.md)
