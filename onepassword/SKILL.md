#### Hazards

**This command prints secrets to stdout.** `get` runs `op item get --format json`,
which includes the item's secret field values, so its output lands in the shell
transcript, in any log capturing it, and in an agent's context. Prefer having the
provider resolve a specific field for another command over reading a whole item, and
never echo the result back. `download` writes a document to the path given, creating
parent directories — confirm that path is git-ignored before using it.

`auth` — and a bare `op`, which runs the same thing — signs in **interactively**. It
prompts for the account password on the terminal and cannot be completed by an agent.
It also **prints the resulting session token into the shell output** as an
`export OP_SESSION_<account>=…` line, so that transcript contains a live credential;
and when `tokenFilename` is configured it writes the same token to that file. Ask a
human to sign in rather than retrying. In CI or unattended use, set
`OP_SERVICE_ACCOUNT_TOKEN`, or `OP_CONNECT_TOKEN` plus `OP_CONNECT_HOST` — either is
detected first and skips the interactive path entirely.

`register` runs `op account add` and is also interactive, prompting for the secret key
and password. It hardcodes `.1password.eu` as the sign-in address, so it only works for
EU-hosted accounts; a `.com` account has to be added with `op account add` directly.
Being an account-level change to the user's local `op` configuration, it affects every
project on the machine, not just this checkout.

Nothing here deletes or edits vault contents: there is no verb that writes to
1Password. The root carries an `Execute`, but this is **not** a passthrough — it
dispatches to `auth` rather than forwarding, so a bare `op` signs you in rather than
printing help, and the rest of the `op` CLI is unreachable from this tree.

**Other commands depend on this one, so its failures surface elsewhere.** Fourteen
providers in this repo take a `*OnePassword` and resolve credentials through it —
`usebruno/bruno`, `grafana/k6`, `gruntwork-io/terragrunt`, `google/gcloud`,
`golang-migrate/migrate`, `foomo/{beam,sesamy,gocontentful}`, `arbitrary/{open,zip}`,
`slack-go/slack`, `webdriverio/webdriverio` and both `pulumi` providers. An expired or
absent session therefore appears as a failure in *those* commands rather than here, and
often as a confusing one, since several of them report it as a missing config value. If
any command fails resolving a secret, check this session before investigating that
command. Note `rclone` also needs `op` but invokes the binary directly rather than
through this provider.

#### Behaviour

Authentication is checked in a fixed order, and only the last is interactive:
`OP_SERVICE_ACCOUNT_TOKEN`; then `OP_CONNECT_TOKEN` + `OP_CONNECT_HOST`; then a
successful `op account get`, which succeeds outright when the desktop app's CLI
integration is enabled; then the cached session. So on a developer machine with the
1Password app integration on, `auth` reports "Already signed in" without ever having
run a sign-in.

The authenticated result is memoised for **10 minutes**, so a session revoked or
expired inside that window still reads as signed in — including in the prompt's status
indicator, which shells out to `op whoami` separately and so can disagree with it.

`tokenFilename`, when set, is both written on sign-in and **reloaded with
`godotenv.Overload`** on each check, which overwrites the current process
`OP_SESSION_<account>`. That is what lets a session survive a shell restart, and it
means a stale file can silently replace a newer in-process session.

Secrets are addressed as `op://<vault>/<item>/<field>` in the terms of the shared
`Secret` type — account, vault, item, field — and every consuming provider embeds that
type into its own config, which is why an undocumented field here propagates into
their schemas. Two distinct template mechanisms exist across the repo and they are not
interchangeable: this provider's own `Render`/`RenderFileTo` use Go `text/template`
with `<% %>` delimiters and an `op` function, while `bruno`, `k6`, `rclone` and both
`pulumi` providers shell out to `op inject` with 1Password's own `{{ op://… }}` syntax.
Check which one a given template file is written for.

The command is named **`op`** by default, not `onepassword`; `CommandWithName` renames
it. `get` and `download` forward `AdditionalArgs()`, so extra `op` flags can go after
`--`; `--account` and `--format json` are always supplied by the provider.

#### Configuration

Read under the `onePassword` key by default — note the camel case, which differs from
both the package name and the command name. `WithConfigKey` renames it and is an
`Option` on `onepassword.New`, **not** a `CommandOption` on `NewCommand`, which only
takes `CommandWithName`. Confirm against the project's own config file. Field shapes:
[`onepassword/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/onepassword/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`account` is what every secret in the project resolves against, so it is worth reading
before diagnosing a resolution failure. `tokenFilename` holds a live session token when
set.

#### Examples

```bash
# Sign in. Interactive — needs a human at the terminal.
posh execute op auth

# Read an item. Prints secret field values as JSON; avoid unless required.
posh execute op get "My Item"

# Save a document to a git-ignored path.
posh execute op download "My Doc" .posh/tmp/doc.pdf
```

#### References

- [`op` CLI reference](https://developer.1password.com/docs/cli/reference/)
- [Service accounts](https://developer.1password.com/docs/service-accounts/) — the non-interactive auth path
- [Secret reference syntax](https://developer.1password.com/docs/cli/secret-reference-syntax/) — the `op://` form the `Secret` type mirrors
- [Provider README](https://github.com/foomo/posh-providers/blob/main/onepassword/README.md) — its plugin snippet does not compile and its help block lists a `signin` verb this tree calls `auth`
