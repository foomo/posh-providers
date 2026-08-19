#### Hazards

Every verb here needs a human, and none of them can be completed by an agent.

`auth` — and a bare `harbor`, which runs the same thing — does nothing but **open a
browser** at the configured login page. It writes no credential and reports no
result: the launcher only spawns the process (`Start()`, never `Wait()`), so a zero
exit status means "a browser was handed the URL", not "you are signed in" and not
even "a browser opened". On a headless host `xdg-open` may well succeed at nothing.
The one error it does return is for an unset or unparseable `authUrl`. There is no
non-interactive Harbor login in this tree.

`docker` **logs the host's docker daemon in**, so the credential it stores lands in
the user's docker config and applies to every other project on the machine, not just
this checkout. It also blocks on an interactive password prompt: it prints the
registry and username, opens the login page in a browser, and then runs
`docker login`, which reads the Harbor **CLI secret** from the terminal. That secret
has to be copied by hand from the Harbor profile page, so an agent cannot supply it.
Ask a human to run this verb; do not retry it.

There is no logout verb, no way to scope the login to the project, and nothing that
undoes `docker`. Neither verb touches Harbor's contents — nothing here pushes,
pulls, deletes or scans an artifact — so the blast radius is credentials and
machine-level docker state rather than registry data.

The root carries an `Execute`, but this is **not** a passthrough: it dispatches to
`auth` rather than forwarding unmatched arguments, and there is no `harbor` binary
involved anywhere. Only the two listed verbs exist, and the rest of Harbor's API is
unreachable from here.

#### Behaviour

`docker` guesses the username before prompting, and the fallback is broken. It first
calls the **GitHub** API with `GITHUB_TOKEN` from the environment and uses that
account's login as the Harbor username — so the identity comes from GitHub, not from
Harbor or from this config, which only makes sense where Harbor authenticates
through GitHub OIDC. When that call fails it prompts on stdin instead, and the helper
it uses returns the input **including the trailing newline**, which is never
trimmed. Since posh joins arguments into a single `sh -c` line, that newline
terminates the line rather than separating two words: anything the provider appends
after the username becomes a *second* shell command. With nothing appended the
username is last and the login still works, so this only surfaces when arguments are
forwarded — `harbor docker -- --password-stdin` reports `sh: 2: --password-stdin: not
found` instead of a login result. Prefer having `GITHUB_TOKEN` set so the prompt path
is never taken.

Forwarding is of limited use on this verb anyway: posh keeps the `--` separator, so it
too is handed to `docker login` as an argument. Both forwarded slots go to that one
command; nothing reaches `auth`, which takes no arguments at all.

The prompt's authentication indicator is inferred from a **deliberately failing
`docker pull`**. `AuthChecker` pulls `<registry>/<project>/null:null` and reads
`Error response from daemon: unknown` — meaning the registry answered and rejected a
missing artifact — as proof of being signed in. Anything else counts as
unauthenticated, so an unreachable registry, a DNS failure or a stopped docker daemon
all render as `Unauthenticated` rather than as an error. Do not read that indicator
as a statement about credentials specifically. Success is also cached in-process for
**one hour**, so a credential that expires or is revoked inside that window still
shows as authenticated.

The registry host is derived from `url` by stripping a `https://` prefix and nothing
else. An `http://` URL keeps its scheme, and a trailing slash survives, either of
which produces a malformed image reference in the auth check. Note the two verbs
disagree about which form to use: `docker login` receives `url` **with** the scheme
while the auth check uses the stripped host.

#### Configuration

Read under the `harbor` key by default. The override is
`harbor.CommandWithConfigKey`, and despite the name it is an `Option` on `harbor.New`
— **not** a `CommandOption` on `NewCommand`; passing it to the latter will not
compile. `CommandWithName` renames the command itself. Confirm against the project's
own config file. Field shapes:
[`goharbor/harbor/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/goharbor/harbor/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`project` is read only by the auth check, never by `auth` or `docker`, so a wrong
value breaks the prompt indicator without affecting either command.

#### Examples

```bash
# Open the Harbor login page. Needs a browser and a human.
posh execute harbor auth

# Log the local docker daemon in. Prompts for the CLI secret
# from the Harbor profile page; affects the whole machine.
posh execute harbor docker
```

#### References

- [Harbor documentation](https://goharbor.io/docs/)
- [Harbor CLI secret](https://goharbor.io/docs/2.0.0/working-with-projects/working-with-images/pulling-pushing-images/)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/goharbor/harbor/README.md) — its plugin snippet calls `harbor.New(l, inst.harbor)` and does not compile; the command constructor is `harbor.NewCommand`
