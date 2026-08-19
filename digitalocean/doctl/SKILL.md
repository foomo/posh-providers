#### Hazards

`auth init` is interactive: it prompts on stdin for a context name and then an API
token, and cannot be completed unattended. It writes that token in plaintext to the
configured config path, inside the checkout — confirm the path is git-ignored before
running it, and never commit the result.

Upstream does skip the prompt when a token arrives via the global `--access-token`,
but do not reach for `-- --access-token <t>`: posh keeps the `--` when forwarding, so
cobra reads what follows as a *positional* and the flag never binds. Since `auth
init` declares no argument count, that fails silently — it prompts anyway rather
than erroring, which is the confusing direction. The reliable non-interactive route
is `DIGITALOCEAN_ACCESS_TOKEN` in the environment, which doctl reads directly. Note
this provider does **not** set that variable for you; see Behaviour.

`registry login` and `registry logout` modify state **outside the project**: they
log the host's Docker daemon in and out of the registry, so the credential lands in
the user's Docker config rather than in the checkout. `logout` therefore breaks any
other work on the machine that was pulling from that registry, and neither verb is
scoped to the posh session. `login` accepts `--never-expire`, which mints an API
token that **never expires** where the default expires after 30 days; that is a
long-lived credential on a shared machine, so treat it as needing approval and
prefer the default.

`kubeconfig` mutates the kubeconfig it writes to. Upstream `kubeconfig save`
**merges** into the target file and, by default, **switches the current context** to
the cluster just saved, so a later bare `kubectl` against that file talks to whatever
was saved last. The provider passes neither `--set-current-context=false` nor
`--alias`, and — per Behaviour — no upstream flag can reach this verb, so the
switch cannot be turned off from posh at all. Each alias normally gets its own file,
which confines the effect; a `CommandWithClusterNameFn` mapping two aliases onto one
filename makes the last write win for both.

The rest of the DigitalOcean CLI is **not** reachable. The root carries no
`Execute`, so this is not a passthrough: only `auth`, `registry` and `kubernetes`
and their listed children exist, and nothing here creates, resizes or deletes
droplets, clusters or registries. There is no destructive cloud verb in this tree.

#### Behaviour

The generated kubeconfig contains **no static credential**. Upstream writes a
client-go `exec` block that shells out to `doctl kubernetes cluster kubeconfig
exec-credential` to fetch a token on demand (renewed as needed, 7 days by default;
the `--expiry-seconds` flag that would bound it is not wired into this tree). Two
consequences worth knowing: an accidentally committed kubeconfig is far less
dangerous than it looks, and the file is **useless anywhere `doctl` is not on
PATH** — copying it into a CI image that lacks the binary produces an
authentication failure, not a permissions error.

That credential plugin also requires the cluster **UUID**, not its name, which is
why the config's `name` field is normally a UUID rather than something readable.
`kubeconfig save` itself accepts either, so a config using human names works for
this command and can still produce a kubeconfig whose plugin invocation is wrong.

Which DigitalOcean *account* a command hits is decided by the doctl config file,
not by anything in this tree. The provider exports `DIGITALOCEAN_CONFIG` at shell
startup, pointing at the configured path, so the session is pinned to that file's
default context. It does **not** export `DIGITALOCEAN_ACCESS_TOKEN`: the code that
would do so is guarded by `err != nil && value != ""`, and since the token getter
returns an empty string on every error path, that condition can never be true — so
the variable is never set from config, and an ambient one from the environment is
left in place. Set it yourself if you need to pin the account explicitly.

`<cluster>` completes from the config map keys, **not** from the DigitalOcean API,
so the list reflects what the project declared rather than what exists in the
account. A cluster that exists remotely but is absent from the config cannot be
named. Completion order is also unsorted (`lo.Keys` over a map with no sort), so the
suggestion order changes between invocations — unlike the sibling `stackit`
provider, which sorts. An unknown alias is rejected before anything runs.

**No upstream flag can be delivered to `kubeconfig`.** It forwards only
`AdditionalArgs`, not `Flags`, so a flag typed on the verb is dropped outright; and
`--` does not rescue it, because posh keeps the separator, cobra then reads the rest
as positionals, and upstream's `kubeconfig save` takes its cluster from `Args[0]`
(already supplied by the provider) while **silently ignoring** any further
positional. So `kubeconfig prod -- --set-current-context=false` neither applies nor
errors — it is discarded without a warning, which is the worst of the three
outcomes. Treat `--set-current-context`, `--alias` and `--expiry-seconds` as
unreachable through posh; run `doctl` directly if you need them. The other verbs go
through a shared helper that forwards `Flags()` normally, so `registry login
--never-expire` does work.

The `--profile` flag is declared on the *internal* flag set and consumed by posh to
choose the output path — it is never passed to doctl.

The godo API client (`Client`, `Cluster`, `Clusters`) and the
`CommandWithAuthTokenProvider` option are present in the package but reached by
nothing in this tree; `authTokenProvider` is assigned and never read. Do not expect
that option to change what any command does.

#### Configuration

Read under the `doctl` key by default. The override is
`doctl.CommandWithConfigKey`, and despite the name it is an `Option` on `doctl.New`
— **not** a `CommandOption` on `NewCommand`; passing it to the latter will not
compile. `CommandWithName` renames the command itself. Confirm against the
project's own config file. Field shapes:
[`digitalocean/doctl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/digitalocean/doctl/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`clusters` is the whole addressable surface — read it to learn which aliases exist
and which cluster ID each maps to. Where the kubeconfig is written is governed by
the `kubectl` provider's own `configPath`, not by anything here.

#### Examples

```bash
# Authenticate first; prompts for a token and needs a human.
posh execute doctl auth init

# Save credentials for the "prod" alias under the default "digitalocean" profile.
posh execute doctl kubernetes kubeconfig prod

# Store it under a different kubectl profile.
posh execute doctl kubernetes kubeconfig prod --profile ci

# Docker registry credentials; affects the host, not the project.
# Flags do work on this verb, unlike on kubeconfig.
posh execute doctl registry login
```

#### References

- [`doctl kubernetes cluster kubeconfig save`](https://docs.digitalocean.com/reference/doctl/reference/kubernetes/cluster/kubeconfig/save/)
- [`doctl auth init`](https://docs.digitalocean.com/reference/doctl/reference/auth/init/)
- [`doctl registry login`](https://docs.digitalocean.com/reference/doctl/reference/registry/login/)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/digitalocean/doctl/README.md)
