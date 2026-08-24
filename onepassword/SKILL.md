#### Hazards

**This command prints secrets to stdout.** `get` fetches an item as JSON *including its
secret field values*, so they land in the transcript and in an agent's context. Prefer
letting another provider resolve the one field it needs, and never echo the result back.
`download` writes a document to the path given — confirm it is git-ignored.

`auth`, and a bare `{{cmd}}` which runs the same thing, signs in **interactively** and
cannot be completed by an agent. It also **prints the session token** as an
`export OP_SESSION_<account>=…` line, so the transcript holds a live credential; with
`tokenFilename` set it writes that token to disk too. Ask a human rather than retrying.
Unattended, set `OP_SERVICE_ACCOUNT_TOKEN`, or `OP_CONNECT_TOKEN` plus
`OP_CONNECT_HOST` — either is detected first and skips the interactive path.

`register` is also interactive and hardcodes `.1password.eu`, so it works only for
EU-hosted accounts, and it changes machine-level account config affecting every project
there.

No verb writes to a vault; nothing edits or deletes vault contents.

**Failures here surface elsewhere.** Fifteen providers in this repo resolve credentials
through this one, so a lapsed session appears as a failure in *those* commands, often
reported as a missing config value. If any command fails resolving a secret, check this
session first.

#### Behaviour

Auth is checked in a fixed order and only the last is interactive: service-account
token; Connect token plus host; an account lookup, which succeeds outright when the
desktop app's CLI integration is on; then the cached session. So with that integration
enabled, `auth` reports "Already signed in" without signing in.

Success is memoised for **10 minutes**, so a session revoked inside that window still
reads as signed in.

`tokenFilename` is reloaded on every check and overwrites the in-process session
variable — that is what survives a shell restart, and it means a stale file can silently
replace a newer session.

Two template mechanisms exist here and are not interchangeable: this provider renders Go
templates with `<% %>` delimiters, while `bruno`, `k6`, `rclone` and both `pulumi`
providers shell out to `inject` with 1Password's own reference syntax.

#### Configuration

Read under the `onePassword` key — camel case, unlike both the package and the command
name. `account` is what every secret resolves against, so read it before diagnosing a
resolution failure.

Field shapes: [`onepassword/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/onepassword/config.schema.json).

#### References

- [CLI reference](https://developer.1password.com/docs/cli/reference/)
- [Service accounts](https://developer.1password.com/docs/service-accounts/) — the non-interactive path
