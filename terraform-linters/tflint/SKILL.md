#### Hazards

**`--fix` rewrites your terraform files in place.** Without it this command is a
read-only linter; with it, the binary edits the `.tf` files in every directory it
processes. No diff, no confirmation, no backup — run it on a clean working tree where
`git diff` can show what changed.

**Omitting the path lints every terraform module in the project**, and that is what
`--fix` then applies to — so one `--fix` with no path can rewrite terraform across the
whole repo. The argument is optional and repeatable; with none given the provider walks
the working directory for every `main.tf`.

**Exit status is the finding, and the loop stops at the first one.** The binary exits
non-zero on issues and the provider returns that error immediately, so directories
after it are never linted. A failure means "issues found here, the rest is unchecked" —
a partial result. With `--fix`, earlier directories may already be rewritten.

**This command is also an `arbitrary/lint` linter.** It implements the interface
`arbitrary/lint` discovers by type assertion, so it runs in any `lint` invocation with
no explicit registration — and `lint --fix` forwards the flag, so that command can
rewrite terraform too. Unlike `tenable/terrascan`, this provider honours `fix`.

#### Behaviour

Discovery is by marker file: directories containing a `main.tf`. The *directory* is
what the binary runs in — the provider sets the working directory rather than passing a
path — so a path should name a directory, and a module without a `main.tf` is invisible
to both completion and the default sweep.

The list skips `node_modules` and dot-directories, and is cached for the session, so a
module added after the shell started is not linted until the cache is cleared.

Only `--fix` is forwarded, and only when set. Neither additional args nor flags after
a `--` separator reach the binary, so upstream options are unreachable. A `.tflint.hcl`
in the linted directory is how rules are configured.

#### Configuration

**This provider has no config key of its own** — no `Config` type and no schema, so
nothing to set in the project's posh config. What gets linted is decided by which
directories contain a `main.tf`; how it is linted, by `.tflint.hcl`.

#### Examples

```bash
posh execute {{cmd}} path/to/module
posh execute {{cmd}} --fix
```

#### References

- [tflint](https://github.com/terraform-linters/tflint) — exit codes and `.tflint.hcl`
- [`arbitrary/lint`](https://github.com/foomo/posh-providers/blob/main/arbitrary/lint/README.md) — the aggregator that runs this, and forwards `--fix`
