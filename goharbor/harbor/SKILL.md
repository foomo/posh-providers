#### Hazards

Every verb here needs a human; an agent cannot complete either.

`auth`, and a bare `{{cmd}}` which runs the same thing, does nothing but **open a
browser** at the configured login page. It writes no credential and reports no result:
the launcher never waits, so a zero exit means "a browser was handed the URL", not "you
are signed in" — on a headless host it succeeds at nothing. There is no non-interactive
login in this tree.

`docker` **logs the host's docker daemon in**, so the credential lands in the user's
docker config and applies to every project on the machine. It also blocks on an
interactive prompt: it prints the registry and username, opens the login page, then runs
the login, which reads the **CLI secret** from the terminal. That secret must be copied
by hand from the Harbor profile page, so an agent cannot supply it. Ask a human; do not
retry.

There is no logout verb and nothing that undoes the login. Neither verb touches registry
contents, so the blast radius is credentials and machine-level docker state.

#### Behaviour

`docker` guesses the username before prompting: it calls the **GitHub** API with
`GITHUB_TOKEN` and uses that account's login as the registry username, which only makes
sense where Harbor authenticates through GitHub OIDC. When that call fails it prompts on
stdin.

The prompt's auth indicator is inferred from a **deliberately failing image pull** of a
nonexistent tag: the daemon's `unknown` response means the registry answered and rejected
a missing artifact, read as proof of being signed in. Anything else counts as
unauthenticated, so an unreachable registry, a DNS failure or a stopped daemon all render
as `Unauthenticated` rather than as an error — do not read it as a statement about
credentials. Success is cached in-process for **one hour**, so a revoked credential still
shows as authenticated inside that window.

#### Configuration

Read under the `harbor` key. `project` is read only by the auth check, never by either
verb, so a wrong value breaks the prompt indicator without affecting any command.

Field shapes: [`goharbor/harbor/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/goharbor/harbor/config.schema.json).

#### Examples

```bash
# Open the login page. Needs a browser and a human.
posh execute {{cmd}} auth

# Log the local docker daemon in. Prompts for the CLI secret.
posh execute {{cmd}} docker
```

#### References

- [CLI secret](https://goharbor.io/docs/2.0.0/working-with-projects/working-with-images/pulling-pushing-images/)
- `goharbor/harbor/README.md` — its plugin snippet does not compile; the constructor is `NewCommand`
