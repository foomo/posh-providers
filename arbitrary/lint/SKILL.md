#### Hazards

**This command's blast radius is invisible from its own tree.** It runs no
linter of its own: it collects every registered posh command that happens to
implement a `Lint` method and invokes it. Which linters those are depends
entirely on what the project's plugin registered, so the same command does
different work in different projects, and naming none runs **all** of them.
Read the completion list, or the project's plugin, to find out what will run.

**`--fix` is forwarded to every linter that runs**, and a linter is free to
rewrite files with it: `terraform-linters/tflint` rewrites `.tf` files in place,
while `tenable/terrascan` and `vsiakka/gherkin-lint` ignore the flag. The files
at risk are the union of whatever the registered linters touch, none of it named
in this command's usage. Run it on a clean tree, and name one linter when fixing.

**Every linter runs concurrently, unlimited, sharing one context**, and the
first failure cancels the rest. A failing run therefore leaves an arbitrary
subset completed - with `--fix`, an arbitrary subset of files already rewritten.
Interleaved output also makes a message hard to attribute to its source.

#### Behaviour

Registration is implicit and by type assertion: any command in the project's
registry with `Name() string` and `Lint(ctx context.Context, fix bool) error`
becomes a linter here. Nothing lists them in configuration, and a provider's own
documentation may not mention it doubles as one, so adding an unrelated provider
can change what this command does.

Naming linters filters the set, matching each linter's `Name()`. A name matching
nothing is warned about individually (`unknown linter: <name>`) and the linters
that did match still run; only when *no* name matches does the command fail.

A failing linter's exit status propagates, so this is usable as a CI gate,
subject to the cancellation caveat above.

**The README is out of date**: the seven `CommandWithGo`-style options it lists
do not exist. `CommandWithName` is the only one.

#### Configuration

**No config key of its own** - no `Config` type, no schema, nothing to set in
the project's posh config. What runs comes from the plugin's command registry;
how each linter behaves comes from that linter's own configuration file.

#### Examples

```bash
posh execute {{cmd}}
posh execute {{cmd}} tflint terrascan
posh execute {{cmd}} tflint --fix
```

#### References

- [`terraform-linters/tflint`](https://github.com/foomo/posh-providers/blob/main/terraform-linters/tflint/README.md) - honours `--fix`, rewrites `.tf` files
- [Provider README](https://github.com/foomo/posh-providers/blob/main/arbitrary/lint/README.md) - its option list does not exist
