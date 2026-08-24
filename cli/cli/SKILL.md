#### Hazards

`auth refresh` **must not be run unattended.** The upstream `gh` prompts on the
TTY and prints a one-time device code a human has to paste into a browser, so an
agent invoking it hangs rather than authenticating. `--reset-token` additionally
discards the currently stored credential, so an abandoned refresh can leave the
shell less authenticated than it started. An agent that finds a missing scope
should report it and the exact `{{cmd}} auth refresh --scopes <scope>` line for
the user to run, not attempt the refresh itself.

`auth status` is read-only and safe for orientation - it reports which hostnames
are authenticated and which token scopes are granted.

Authentication is process-wide, not per-command. `CLI.LoadToken` shells out to
the binary's `auth token` verb and exports the result as `GITHUB_TOKEN` for the whole posh
process, so any other command in the project reading `GITHUB_TOKEN` depends on
being authenticated here first. It is a no-op when `GITHUB_TOKEN` is already set
and gives up after one second. `LoadToken` is called by the project's plugin,
not by this tree - so a `GITHUB_TOKEN` failure elsewhere is diagnosed here even
though nothing in this tree sets the variable.

#### Behaviour

The root node has no `Execute`, so - unlike most providers here - the rendered
subtree *is* the whole surface. `gh` verbs outside `auth status` and
`auth refresh` are not reachable through this command unless the project
exposes the binary another way.

If the project registers `ScopesChecker`, the prompt runs
`gh` with `auth status --hostname <host>` per configured host each tick, flagging
hosts whose granted scopes do not cover the configured ones. A red `gh:` entry
in the prompt is that check, and its remedy is the interactive refresh above.

#### Configuration

The `gh` key's `scopes` maps a `gh` hostname to the OAuth scopes the project
requires. It only feeds `ScopesChecker` and does not constrain what the two
subcommands accept.

Field shapes: [`cli/cli/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/cli/cli/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} auth status
```

#### References

- [GitHub CLI manual](https://cli.github.com/manual/)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/cli/cli/README.md)
