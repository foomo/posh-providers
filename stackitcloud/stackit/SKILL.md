#### Hazards

Both verbs are agent-hostile, and neither can be completed unattended.

`auth` runs `stackit auth login`, which opens a **browser window** for a web
authorization flow. There is no flag that turns this into a headless login, so an
agent invoking it will hang on a browser it cannot drive. The non-interactive
path exists but is not wired into this tree: run `stackit auth
activate-service-account` outside posh (with `--service-account-key-path` or
`--service-account-token`, or `--use-oidc` for workload identity), and every
command here then works against that session. Ask a human to authenticate rather
than retrying `auth`.

`kubeconfig` **cannot be run unattended at all**, and the usual `--` escape does
not help. Upstream prompts for confirmation before writing the file, and the
global `--assume-yes`/`-y` flag that would skip it cannot be delivered through
this command: posh forwards `AdditionalArgs` with the literal `--` still attached,
so anything typed after it arrives as a *positional* argument, and `ske kubeconfig
create` accepts exactly one (the cluster name). So
`kubeconfig dev -- --assume-yes` fails with an "expected exactly 1 argument" error
rather than auto-confirming. Treat this verb as requiring a human at the terminal.

`kubeconfig` writes into the checkout: the kubectl config path is resolved against
the project root, so the file lands **inside the repository**. Confirm the path is
git-ignored before running it. The provider does not pass `--login`, so the file
upstream writes is the *dynamic* kind — it embeds no secret and instead shells out
to the STACKIT CLI for short-lived credentials on demand. That makes an accidental
commit far less damaging than it looks, but it is still a file naming a real
cluster, and anyone who has it plus a valid STACKIT session can use it. Note the
flip side: because the credentials come from the CLI at use time, the kubeconfig
only works where the CLI is installed and logged in, so it is not portable to CI by
copying. Should a project add `--login` (via `--`, which does reach this verb for
non-flag tokens), the file becomes static secrets plus the account's email address,
and revoking it then requires a full SKE credential rotation that invalidates every
other admin kubeconfig for the cluster.

Nothing here deletes or reconfigures cloud resources: there is no cluster
create/delete verb, and `kubeconfig` only mints credentials. The whole tree is a
two-verb read/auth surface, and the root has no `Execute`, so unlisted `stackit`
subcommands are **not** reachable — this is not a passthrough, and the rest of the
STACKIT CLI is deliberately out of reach.

#### Behaviour

What expires is the **CLI login session, not the file**. Because the generated
kubeconfig fetches credentials through the CLI on each use rather than embedding
them, it does not go stale on its own — but the login it depends on defaults to 2
hours (adjustable with `stackit config set --session-time-limit`, which this tree
does not expose). So a `kubectl` call that worked earlier can start failing on
authentication with the kubeconfig unchanged and the cluster perfectly healthy.
Read that failure as an expired login needing `auth` again, not as a broken
kubeconfig to regenerate. Upstream also has an `--expiration` flag (default 1h)
that the provider never passes; it bounds the credentials the file requests, not
the file itself.

The provider passes no `--overwrite`, so upstream **merges** into the target file
rather than replacing it. Each alias gets its own file (`<alias>.yaml`), so this is
usually invisible — but re-running for the same alias updates that context in place
instead of rewriting the file, and if a project's `CommandWithClusterNameFn` maps
two aliases onto one filename, both accumulate as separate contexts in a single
kubeconfig. In that case the file's *current context* decides which cluster
`kubectl` talks to, and this command never sets it.

`<project>` and `<cluster>` are both **local aliases, not STACKIT names**. Each
completes from the config map keys, and the real values live in the entries: the
project's `id` is a UUID passed as `--project-id`, and the cluster's `name` is the
actual SKE cluster passed as the positional argument. So the cluster you type is
not necessarily the cluster the CLI receives — `clusters: {dev: {name:
my-project-dev-cluster}}` means typing `dev` acts on `my-project-dev-cluster`.
Read the config to learn the mapping; nothing in the tree or the completion shows
it.

The kubeconfig **filename** is derived from the alias, not from the real cluster
name: it is written to `<kubectl configPath>/<profile>/<alias>.yaml`, with
`profile` defaulting to `stackit`. A project can override the filename derivation
with `CommandWithClusterNameFn`, which receives both the alias and the resolved
`Cluster`, so in a project that passes it the filename may match neither. Nothing
validates that the resulting path is one `kubectl` will later look for.

An unknown project or cluster alias is rejected with `given project not found` /
`given cluster not found` before anything runs, so a typo is safe here — unlike
several sibling cluster providers in this repo, which construct a kubectl cluster
for any string.

#### Configuration

Read under the `stackit` key by default. The override is
`stackit.CommandWithConfigKey`, and despite the name it is an `Option` on
`stackit.New` — **not** a `CommandOption` on `NewCommand`; passing it to the
latter will not compile. `CommandWithName` renames the command itself. Confirm
against the project's own config file. Field shapes:
[`stackitcloud/stackit/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/stackitcloud/stackit/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`projects` is the whole addressable surface — a project or cluster absent from it
cannot be reached by any argument. Read it to learn which aliases exist and which
real cluster each one maps to. Where the kubeconfig is written is governed by the
`kubectl` provider's own `configPath`, not by anything here.

#### Examples

```bash
# Authenticate first; opens a browser and needs a human.
posh execute stackit auth

# Write a kubeconfig for the "dev" alias of project "my-project".
# Prompts for confirmation; cannot be automated.
posh execute stackit project my-project kubeconfig dev

# Store it under a named kubectl profile instead of the default "stackit".
posh execute stackit project my-project kubeconfig dev --profile ci
```

#### References

- [STACKIT CLI user guide](https://docs.stackit.cloud/developer-tools/stackit-cli/)
- [`stackit ske` command reference](https://github.com/stackitcloud/stackit-cli/blob/main/docs/stackit_ske.md)
- [`stackit auth activate-service-account`](https://github.com/stackitcloud/stackit-cli/blob/main/docs/stackit_auth_activate-service-account.md) — the non-interactive auth path
- [SKE credential rotation](https://docs.stackit.cloud/products/runtime/kubernetes-engine/how-tos/rotate-ske-credentials/)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/stackitcloud/stackit/README.md)
