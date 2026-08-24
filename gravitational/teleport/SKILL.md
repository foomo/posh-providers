#### Hazards

**Every verb here authenticates a human interactively.** Login opens a browser for GitHub
SSO and the other verbs inherit that session. An agent cannot complete any of them
unattended: they block on a browser round-trip or fail with an expired session. Treat
this as something a human runs before the agent starts.

**A bare `{{cmd}}` is not a help screen — it logs in.** The root carries an `Execute`
running the same login as `auth`, so typing the command name alone to see what it offers
starts an SSO flow.

**`kubeconfig` rewrites the profile's kubeconfig.** The upstream login merges into
whatever is at `KUBECONFIG`, so the provider moves the old file aside first, restoring it
if the login fails. It is still replaced on success, so anything hand-added is lost.

**The GitHub connector is hardcoded**, so a cluster using a different one cannot log in
here regardless of config.

An ambiguous cluster alias is rejected rather than resolved: two clusters sharing an
alias, or an alias equal to another cluster's real name, error and name the conflict.
Fixing it is a config edit.

#### Behaviour

**Nothing is listed until you are authenticated.** The cluster, app and database listings
all return nil when the status check fails, so before login those arguments complete to
nothing — indistinguishable from "this project has none".

**Discovery is filtered by the configured labels, silently.** All three listings build a
query from `labels` as one `&&` expression, so anything missing a configured label is
invisible to completion; empty `labels` lists everything. If a cluster you know exists is
not offered, check the labels before concluding it is gone.

Listings are memoised for the session, so a resource added upstream afterwards needs a
`cache clear`. The auth check has its own **12-hour** memo, so a session revoked inside
that window still reads as authenticated.

`TELEPORT_HOME` is exported process-wide at construction from `path`, so every upstream
invocation in the shell — including from other providers — uses that session directory
rather than the user's own.

#### Configuration

Read from the `teleport` key. The provider also supplies an auth checker surfacing login
state in the prompt. The README's config sample omits `apps`, which is a real key.

Field shapes: [`gravitational/teleport/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/gravitational/teleport/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} auth
posh execute {{cmd}} kubeconfig dev --profile ci
```

#### References

- [`tsh` CLI reference](https://goteleport.com/docs/reference/cli/tsh/) — the binary every verb wraps
