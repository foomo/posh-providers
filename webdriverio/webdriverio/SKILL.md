#### Hazards

**Omitting `[path]` runs the suite in every discovered directory, serially, and stops at the first
failure.** With a path given, one directory runs. Without one, `execute` loops over everything
`wdio.conf.ts` discovery found and returns on the first non-zero exit — so a partial run leaves
earlier directories tested, the failing one incomplete and later ones never attempted. The usage block
shows `[path]` as optional, which reads as narrower than it is.

**These are end-to-end tests fired at a real, configured environment.** The `<env>` argument selects a
domain from config and the run drives a browser against it, so whatever the suite does — logging in,
writing data, sending mail — happens on that host. Nothing in this provider restricts `<env>` to a
staging target, and there is no dry-run. Confirm which environment is selected before running
unattended.

**A mistyped `<site>` or `<env>` is not rejected; it silently produces an empty base URL.** There is
no `Validate` anywhere in this provider, and `execute` indexes the config maps directly
(`c.cfg.Sites[site][env]`), so an unknown key yields a zero `ConfigEnv` with an empty `Domain`. The run
then proceeds with `E2E_BASE_URL` set to just the mode's host prefix and port, and no basic auth is
resolved. The tests fail confusingly rather than the command refusing the argument. Completion offers
only configured names, so this bites on typed or scripted invocations.

#### Behaviour

**The mode name `browserstack` is a magic string, not a config flag.** `execute` compares
`mode == "browserstack"` literally, so that one key in `modes:` switches the whole run onto
BrowserStack — resolving credentials from 1Password and setting `E2E_ENV=browserstack` plus the
`E2E_PROJECT`/`E2E_BUILD_*` variables. Any other mode name sets `E2E_ENV=chromium`. Renaming that key
in config silently turns the integration off; nothing warns.

**`--ci` takes precedence over the mode and skips credential resolution.** The branch order is `--ci`
first, so `--ci` sets `E2E_ENV=ci` and BrowserStack is never contacted even in `browserstack` mode —
no 1Password lookup, no `BROWSERSTACK_*` variables. Selecting the `browserstack` mode without a
`browserStack:` config errors, naming the missing key.

**Almost everything reaches the runner as an environment variable, not a flag.** `LOG_LEVEL`,
`E2E_ENV`, `E2E_BASE_URL`, `BASIC_AUTH`, `HEADLESS`, `debug`, `SCENARIOS` and the BrowserStack set are
all exported into the child process; only `--spec`/`--suite` and `--tag` become command-line arguments.
So the suite's own `wdio.conf.ts` has to read those variables for the flags to do anything — a flag
that appears to work here does nothing if the project's config ignores the variable. Note `debug` is
exported lowercase while every other variable is upper-case.

**`NODE_TLS_REJECT_UNAUTHORIZED=0` is always set.** The provider disables TLS certificate verification
for every run, not just local ones, so a run against a real environment will not notice an invalid or
substituted certificate.

**The base URL is assembled from two config sections.** `<env>`'s `domain` is the base; the mode's
`hostPrefix` is prepended and its `port` appended, giving
`[<hostPrefix>.]<domain>[:<port>]`. Note no scheme is added — whatever the suite does with
`E2E_BASE_URL` decides that.

**Discovery keys on `wdio.conf.ts` and assumes it sits under `e2e/`.** `paths()` walks for
`wdio.conf.ts` (ignoring dotfiles and `node_modules`), then strips the suffix `/e2e/wdio.conf.ts` to
get the directory. A config that is *not* under an `e2e/` directory does not match that suffix, so the
untrimmed **file path** is returned as if it were a directory — and the `os.Stat` guard does not catch
it, since stat succeeds on a regular file. The run then sets its working directory to a file. Verified
by running the trim over each layout. Keep `wdio.conf.ts` under `e2e/`. The runner is always invoked as
`wdio run e2e/wdio.conf.ts` relative to the chosen directory, which is the same assumption from the
other side.

**Discovery and the completion lists are cached per session.** Paths, specs, tags and scenarios are all
memoised, so a newly added feature file or directory needs a `cache clear` before it is suggested.

**`--tag` and `--scenario` completion needs `[path]` typed first.** Both are populated only when the
argument count reaches four, and they are derived by shelling out to `find | grep` over
`<path>/e2e/features` — so they are empty until a path is given, and they depend on the feature files
using tab-indented `@tag` and `Scenario:` lines.

#### Configuration

Read from the `webdriverio` key by default; `WithConfigKey` overrides it, and `CommandWithName`
renames the command itself (default `wdio`). Note the option is `WithConfigKey`, not
`CommandWithConfigKey` as most siblings spell it.

Two documented fields are **dead**: `dirs` and `secrets` are declared on `Config` and appear in the
schema, but nothing in the provider reads either — directories come from `wdio.conf.ts` discovery, and
the only secrets resolved are `sites.<site>.<env>.auth` and `browserStack`. The README presents `dirs`
as the way to point at test directories; it is not.

- [config.schema.json](https://raw.githubusercontent.com/foomo/posh-providers/main/webdriverio/webdriverio/config.schema.json)

#### Examples

```bash
# one site/env, every discovered directory
posh execute wdio local shop staging

# narrow to a single directory
posh execute wdio local shop staging ./frontend

# a single feature file
posh execute wdio local shop staging ./frontend --spec login.feature

# headless, stopping at the first failure
posh execute wdio local shop staging ./frontend --headless --bail

# BrowserStack, resolving credentials from 1Password
posh execute wdio browserstack shop staging ./frontend

# the same mode on CI, where no credentials are resolved
posh execute wdio browserstack shop staging ./frontend --ci
```

#### References

- [WebdriverIO docs](https://webdriver.io/) — the runner, its config file and its own CLI flags
- [Cucumber tag expressions](https://cucumber.io/docs/cucumber/api/#tag-expressions) — the syntax `--tag` forwards
- [provider README](README.md) — note it documents `dirs` as functional; it is not
