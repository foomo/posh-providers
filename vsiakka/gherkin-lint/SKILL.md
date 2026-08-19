#### Hazards

**This command is read-only** - it lints `.feature` files and reports
violations without modifying anything. There is no `--fix`, and the `fix`
parameter it receives as a linter is ignored. Nothing here needs approval on the
grounds of side effects.

**Exit status is the finding, and the loop stops at the first one.**
gherkin-lint exits non-zero when it reports violations, and the provider returns
that error immediately, so the directories after it are never linted. A failure
means "violations found here, the rest is unchecked" - a partial result, not a
complete report.

**Omitting the path lints every matching directory in the project**, which is
the intended "lint everything" case.

**This command is also an `arbitrary/lint` linter.** It implements the interface
`arbitrary/lint` discovers by type assertion, so it runs as part of any `lint`
invocation with no explicit registration - a much wider action than invoking it
directly. It ignores `lint --fix`, so it never rewrites anything, but be aware
other linters in that same sweep may.

#### Behaviour

**The marker file is `wdio.conf.ts`, not a `.feature` file.** Discovery finds
WebdriverIO project directories, and the *directory* is what gets passed to
gherkin-lint - the tool then finds the feature files beneath it. That is
deliberate rather than a mistake: in these projects the features live at
`<dir>/e2e/features`, as the `webdriverio` provider's own lookup confirms. The
consequences are worth knowing: a directory full of `.feature` files with no
`wdio.conf.ts` beside it is invisible to both completion and the default sweep,
and a path you pass by hand is handed over whether or not it contains features.

Discovery skips `node_modules` and any dot-directory, and is cached for the
session, so a project added after the shell started is not offered or linted
until the cache is cleared.

Only the directory is passed. This tree declares no flags at all, and neither
additional args nor additional flags after a `--` separator reach the binary -
so upstream options like `--config` or `--format` are not reachable from here.
gherkin-lint's own `.gherkin-lintrc` in the target directory is the way to
configure rules.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type, no
`config.schema.json` and no `config.base.json`, so there is nothing to set in
the project's posh config and no entry for it in `posh.schema.json`. What gets
linted is decided by which directories contain a `wdio.conf.ts`; how it is
linted is decided by gherkin-lint's own config file.

The command can be renamed with `CommandWithName`, and the gherkin-lint binary
invocation swapped with `CommandWithExecGherkinLint` - both options on
`NewCommand`.

#### Examples

```bash
# Lint one WebdriverIO project directory
posh execute gherkin-lint path/to/project

# Lint every directory containing a wdio.conf.ts
posh execute gherkin-lint
```

#### References

- [gherkin-lint](https://github.com/vsiakka/gherkin-lint) - the linter this wraps, its exit codes and `.gherkin-lintrc`
- [`arbitrary/lint`](https://github.com/foomo/posh-providers/blob/main/arbitrary/lint/README.md) - the aggregator that also runs this
- [Provider README](https://github.com/foomo/posh-providers/blob/main/vsiakka/gherkin-lint/README.md)
