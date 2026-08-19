#### Hazards

**Every verb here authenticates a human interactively.** `tsh login` opens a browser for GitHub SSO,
and the other verbs inherit that session. An agent cannot complete any of them unattended: they either
block waiting for a browser round-trip or fail with an expired session. Treat this provider as
something a human runs before the agent starts working, not something the agent runs.

**Bare `teleport` is not a help screen — it logs in.** The root node carries `Execute` and runs the
same `tsh login --proxy=<hostname> --auth=github` as `teleport auth`. Typing the command name alone to
see what it offers starts an SSO flow instead.

**`kubeconfig` deletes the existing kubeconfig before it logs in, not after it succeeds.** It calls
`DeleteConfig(<profile>)` first and only then runs `tsh kube login`. If the login fails — expired
session, wrong cluster name, no network — the previous working config is already gone and the profile
is left with no kubeconfig at all. Re-run it after authenticating rather than assuming the old
credentials survived.

**`--auth=github` is hardcoded.** The provider always requests the GitHub connector, so a project
whose teleport cluster uses a different one cannot log in through this command regardless of config.

**Cluster aliases are reverse-mapped by scanning the alias map, so an ambiguous or shadowing alias
sends you to the wrong cluster.** Completion offers the alias, and `kubeconfig` maps it back with
`Kubernetes.Name()`, which iterates the map looking for a matching value. Two consequences, both
verified by running the lookup: if two real clusters share one alias the winner is **random per call**
(a 175/25 split over 200 runs), and if an alias equals some *other* cluster's real name, the alias
wins — configuring `kubernetes-dev: prod` means typing `prod` logs you into the dev cluster. Keep
alias values unique and disjoint from every real cluster name; nothing validates this.

#### Behaviour

**Nothing is listed until you are authenticated.** `Clusters`, `Apps` and `Databases` all return `nil`
when `tsh status` fails, so before login the arguments to `kubeconfig`, `app` and `database` complete
to nothing at all — which looks identical to "this project has none configured". Run `teleport auth`
first.

**Discovery is filtered by the configured labels, and the filter is silent.** All three listings pass
`--query` built from `labels:` as a single `&&` expression, so anything not carrying every configured
label is invisible to completion. An empty `labels:` produces an empty query and lists everything. If a
cluster you know exists is not offered, check the labels before concluding it is gone.

**The listings are cached for the session.** Clusters, apps and databases are memoised under fixed keys
after the first successful call, so a resource added in teleport afterwards needs a `cache clear`. The
authentication check has its own 12-hour memo: once `tsh status` succeeds, `IsAuthenticated` returns
true from memory for 12 hours without re-checking, so a session revoked in the meantime still reads as
authenticated in the prompt.

**`database` takes its user from the environment first.** `--db-user` is `TELEPORT_DATABASE_USER` when
that variable is set, falling back to `database.user` from config. Neither is shown in the usage block.

**`app` prepends per-app arguments from config.** `apps:` maps an app name to extra arguments inserted
before the name in `tsh apps login`, so what a given app does depends on config the tree cannot show.
Apps absent from the map still work; they just get no extra arguments. Note the README omits this key
entirely.

**All four verbs forward flags to `tsh` verbatim.** `r.Flags()`, `AdditionalArgs()` and
`AdditionalFlags()` are all passed on, so any `tsh` flag can be given directly — the tree declares only
`--profile`, on `kubeconfig`, and that one is consumed by posh to pick which kubeconfig to write.

**`TELEPORT_HOME` is set process-wide at construction.** `NewTeleport` exports it from `path:` when the
provider is built, so every `tsh` invocation in the shell — including ones from other providers — uses
that session directory rather than the user's own `~/.tsh`.

#### Configuration

Read from the `teleport` key by default. The override is `CommandWithConfigKey`, which despite the name
is an option on **`NewTeleport`**, not on `NewCommand`; `NewCommand` takes `CommandWithName` to rename
the command itself. The provider is built in two pieces — `NewTeleport(l, cache)` for the client and
`NewCommand(l, cache, teleport, kubectl)` for the command — and it also supplies `AuthChecker`, which
surfaces the login state in the posh prompt.

The README's config sample omits `apps:`, which is a real key.

- [config.schema.json](https://raw.githubusercontent.com/foomo/posh-providers/main/gravitational/teleport/config.schema.json)

#### Examples

```bash
# log in - opens a browser for GitHub SSO
posh execute teleport auth

# the bare command does the same thing
posh execute teleport

# write a kubeconfig for a cluster, by alias
posh execute teleport kubeconfig dev

# ... into a named profile
posh execute teleport kubeconfig dev --profile teleport

# log in to a database, using database.user or TELEPORT_DATABASE_USER
posh execute teleport database orders

# log in to an app, with any configured per-app arguments prepended
posh execute teleport app grafana

# drop the session
posh execute teleport logout
```

#### References

- [tsh CLI reference](https://goteleport.com/docs/reference/cli/tsh/) — the binary every verb wraps
- [provider README](README.md) — note its config sample omits the `apps` key
