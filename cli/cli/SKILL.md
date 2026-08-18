#### Hazards

`gh auth status` is read-only and safe for orientation - it reports which
hostnames are authenticated and which token scopes are granted.

`gh auth refresh` **must not be run unattended.** It is an interactive flow: the
upstream `gh` prompts on the TTY and prints a one-time device code that a human
has to paste into a browser, so an agent invoking it hangs rather than
authenticating. `--reset-token` additionally discards the currently stored
credential, so a failed or abandoned refresh can leave the shell less
authenticated than it started. An agent that finds a missing scope should report
the missing scope and the exact `gh auth refresh --scopes <scope>` line for the
user to run, not attempt the refresh itself.

Both subcommands dispatch to the same `execAuth`, which forwards the matched
path plus any flags to the real `gh` binary. The root node has no `Execute` of
its own, so - unlike most providers here - the rendered subtree *is* the whole
surface: `gh` verbs outside `auth status` and `auth refresh` are not reachable
through this command.

Authentication is process-wide, not per-command. `CLI.LoadToken` shells out to
`gh auth token` and exports the result as `GITHUB_TOKEN` for the whole posh
process, so any other command in the project that reads `GITHUB_TOKEN` depends
on `gh` being authenticated first. It is a no-op when `GITHUB_TOKEN` is already
set, and it gives up after one second. `LoadToken` is called by the project's
plugin, not by this command tree - so a `GITHUB_TOKEN`-related failure elsewhere
is diagnosed here with `gh auth status`, even though nothing in this tree sets
the variable.

If the project registers `ScopesChecker`, the prompt runs
`gh auth status --hostname <host>` for every configured host on each tick and
flags hosts whose granted scopes do not cover the configured ones. A red `gh:`
entry in the prompt is that check, and its remedy is the interactive refresh
above.

#### Configuration

Config key `gh` by default, but the project can rename it via `WithConfigKey`, so
confirm the actual key against the project's own posh config rather than assuming
it. `scopes` maps a gh hostname to the OAuth scopes the project requires; it only
feeds `ScopesChecker`, and does not constrain what the two subcommands accept.

Field shapes: [`cli/cli/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/cli/cli/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same
schema is bundled under in the project's `posh.schema.json`.

#### Examples

```bash
# Report auth state for every configured host
posh execute gh auth status

# The line to hand a user when a scope is missing - do not run it unattended
posh execute gh auth refresh --scopes repo --scopes workflow
```

#### References

- [GitHub CLI manual](https://cli.github.com/manual/)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/cli/cli/README.md)
