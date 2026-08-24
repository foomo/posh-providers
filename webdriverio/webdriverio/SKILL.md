#### Hazards

**These are end-to-end tests fired at a real, configured environment.** `<env>`
selects a domain from config and the run drives a browser against it, so
whatever the suite does - logging in, writing data, sending mail - happens on
that host. Nothing restricts `<env>` to a staging target and there is no dry
run. Confirm which environment is selected before running unattended.

**Omitting `[path]` runs the suite in every discovered directory, serially,
stopping at the first failure**, leaving earlier directories tested, the failing
one incomplete and later ones never attempted. The usage block shows it as
optional, which reads narrower than it is.

**A mistyped `<site>` or `<env>` is not rejected; it silently produces an empty
base URL.** There is no `Validate` here and the config maps are indexed
directly, so an unknown key yields an empty domain and no basic auth - the tests
then fail confusingly instead of the command refusing the argument.

**`NODE_TLS_REJECT_UNAUTHORIZED=0` is always set**, for every run and not just
local ones, so a run ignores an invalid or substituted certificate.

#### Behaviour

**The mode name `browserstack` is a magic string, not a config flag.** It is
compared literally, so that one key in `modes:` switches the whole run onto
BrowserStack, resolving 1Password credentials; any other name sets
`E2E_ENV=chromium`, so renaming it silently turns the integration off. Selecting
it without a `browserStack:` config errors, naming the missing key. `--ci` wins
over the mode.

**Almost everything reaches the runner as an environment variable, not a flag**;
only `--spec`/`--suite` and `--tag` become command-line arguments. So the
project's own `wdio.conf.ts` has to read those variables for the flags to do
anything.

**Discovery assumes the config sits under `e2e/`**, stripping that suffix from
each match, so a config elsewhere leaves the untrimmed **file path** returned as
if it were a directory - and the `os.Stat` guard does not catch it, since stat
succeeds on a file. The run then sets its working directory to a file.

#### Configuration

Read from the `webdriverio` key; the override option is spelled
`WithConfigKey`, not `CommandWithConfigKey` as most siblings do.

Two fields are **dead**: `dirs` and `secrets` appear in the schema but nothing
reads either - the only secrets resolved are `sites.<site>.<env>.auth` and
`browserStack`. The README presents `dirs` as the way to point at test
directories; it is not.

Field shapes: [`config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/webdriverio/webdriverio/config.schema.json).

#### References

- [Upstream runner docs](https://webdriver.io/)
- [Cucumber tag expressions](https://cucumber.io/docs/cucumber/api/#tag-expressions)
