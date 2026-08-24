#### Hazards

**This command is read-only** - it reports violations and modifies nothing.
There is no `--fix`, and the `fix` parameter it receives as a linter is ignored.

**Exit status is the finding, and the loop stops at the first one.** The linter
exits non-zero when it reports violations and the provider returns that error
immediately, so directories after it are never linted. A failure means
"violations found here, the rest is unchecked" - a partial result, not a
complete report.

**Omitting the path lints every matching directory in the project**, which is
the intended "lint everything" case.

**This command is also an `arbitrary/lint` linter.** It implements the interface
`arbitrary/lint` discovers by type assertion, so it runs as part of any project
level lint sweep with no explicit registration - a much wider action than
invoking it directly. It ignores that sweep's `--fix`, but other linters in the
same sweep may rewrite files.

#### Behaviour

**The marker file is `wdio.conf.ts`, not a `.feature` file.** Discovery finds
WebdriverIO project directories and passes the *directory* to the linter, which
then finds the feature files beneath it - deliberate, since in these projects
features live at `<dir>/e2e/features`. Consequences: a directory full of
`.feature` files with no `wdio.conf.ts` beside it is invisible to both
completion and the default sweep, and a path passed by hand is handed over
whether or not it contains features.

Discovery skips `node_modules` and any dot-directory, and is cached for the
session, so a project added after the shell started is not offered or linted
until the cache is cleared.

Only the directory is passed. This tree declares no flags, and neither
additional args nor anything after a `--` separator reaches the binary, so
upstream options like `--config` or `--format` are not reachable from here. The
`.gherkin-lintrc` in the target directory is the way to configure rules.

#### Configuration

No config key, no schema - nothing to set in the project's posh config. What
gets linted is decided by which directories contain a `wdio.conf.ts`; how it is
linted by the linter's own config file.

#### References

- [Upstream linter](https://github.com/vsiakka/gherkin-lint) - exit codes and `.gherkin-lintrc`
- [`arbitrary/lint`](https://github.com/foomo/posh-providers/blob/main/arbitrary/lint/README.md) - the aggregator that also runs this
- [Provider README](https://github.com/foomo/posh-providers/blob/main/vsiakka/gherkin-lint/README.md)
