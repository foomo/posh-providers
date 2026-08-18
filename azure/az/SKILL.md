#### Hazards

The root node carries no `Execute`, so unlike most providers here this tree
*is* the surface: an unrecognised first argument is rejected rather than
forwarded. The passthrough is one explicit subcommand instead - `az raw`, which
drops the word `raw` and hands everything after it to the real `az` binary. So
`az raw group delete --name x` reaches the upstream CLI unchanged, with every
side effect it has and no confirmation step from posh. Treat `raw` as the whole
Azure CLI, not as a listed leaf; only run mutating verbs through it with
approval. `az raw` with nothing after it runs a bare `az`.

Every invocation of every subcommand runs with `AZURE_CONFIG_DIR` pointed at
this provider's configured path. That is set once when the shell starts, so the
signed-in account, tokens and subscription cache live per project rather than in
the user's real home directory - an account logged in outside posh is invisible
here, and one logged in here does not leak out.

`login` needs a browser unless it is given a service principal. With
`--service-principal <name>` it does a non-interactive credential login, reading
the client id, secret and tenant straight out of the posh config - so it is the
only variant an agent can complete unattended, and it puts a plaintext secret
from the config onto the `az` command line. Without the flag it runs
`az login --allow-no-subscriptions --tenant <tenantId>`, which opens a browser
and cannot finish without a graphical session; ask a human. The service
principal names come from the `az` config key, not from a live lookup.
`logout` drops those credentials for every subsequent command in the session.

`kubeconfig` mutates cluster access and depends on two things outside this
provider. It runs `aks get-credentials --overwrite-existing`, so it silently
replaces any existing entry for that cluster, then shells out to `kubelogin
convert-kubeconfig -l azurecli`, which must be installed separately - the
provider never checks for it and the failure surfaces as an exec error after the
credentials have already been written. The destination file is
`<kubectl configPath>/[<profile>/]<cluster>.yaml`, owned by the `kubectl`
provider, and it is the prerequisite for every kubectl-based command against
that cluster. Run `login` first: `aks get-credentials` fails on authentication
otherwise. If the cluster config sets `proxyUrl`, the file is rewritten
afterwards to inject it into the current context's cluster entry.

`vault` is the destructive corner of the tree. `key delete`, `secret delete` and
`certificate delete` remove live key vault material; `secret set` and
`certificate set` overwrite an existing entry with no diff, and `key create`
adds a new key version. None prompt. `secret show` and `certificate show` print
the secret value to stdout, so they are read-only but not safe to log. Only
`list` and the `show` verbs are safe unattended, and the target is chosen by the
first two words after `vault` - a local invocation and a production one differ
by one argument.

`artifactory` runs `az acr login`, which writes a registry credential into the
local Docker config - it changes docker state, not just Azure state.

#### Behaviour

`--profile` on `kubeconfig` does not reach `az`. It only selects which
subdirectory of the kubectl provider's config path the credentials land in, so
it partitions kubeconfigs; `azure` is offered as a suggestion but any name is
accepted and becomes a directory. It defaults to empty, writing to the
top-level `<cluster>.yaml`. Whatever you pass must match what the consuming
kubectl-based command is told to read.

A cluster needs no matching entry in the kubectl provider's config for
`kubeconfig` to succeed - the lookup that looks like a guard at `command.go:457`
can never fail, because `kubectl.Cluster` constructs the value rather than
looking it up. The path is derived from the name, so a name accepted by the az
config is written either way and only the az-side lookups reject a typo.

The first two words after `vault` are dynamic tree nodes, not arguments, so the
catalog renders them as `<subscription>` and `<vault>` placeholders. Both resolve
against the `az` config key: the subscription names are its `subscriptions`
keys, and the vault names are the `vaults` keys of the subscription you picked -
so the second is only meaningful after the first. The same holds for
`artifactory` and `kubeconfig`, whose two arguments are the `artifactories` and
`clusters` keys of the chosen subscription. In every case the config key is a
posh-side label; the `name` and `resourceGroup` beneath it are what reach Azure.

The entry `name` of a `show` or `delete` completes from a live
`az keyvault <noun> list` against the vault, cached per subscription and vault -
so tab completion there makes a network call and needs a working login.

`login`, and every `vault` verb, append anything after a `--` separator to the
underlying `az` call, which is how you reach upstream flags the tree does not
model. `artifactory` and `kubeconfig` do not; extra arguments there are ignored.

#### Configuration

The `az` key of this project's posh config holds the tenant, the service
principals and the whole subscription tree that every placeholder in this
command resolves against - read that key for the valid values. `az` is only the
default key, overridable via `WithConfigKey`, so confirm the actual one against
the project's own file.

Field shapes: [`azure/az/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/azure/az/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

Three things the schema cannot express: `configPath` is created on startup and
becomes `AZURE_CONFIG_DIR` for every invocation, a `servicePrincipals` entry is
what decides whether `login` can run unattended, and the `proxy` field is read
by nothing in this provider - only a cluster's `proxyUrl` has an effect.

#### Examples

```bash
posh execute az login --service-principal <name>
posh execute az kubeconfig <subscription> <cluster> --profile azure
posh execute az artifactory <subscription> <artifactory>
posh execute az vault <subscription> <vault> secret list --output json
posh execute az raw account show
```

#### References

- https://learn.microsoft.com/en-us/cli/azure/reference-index
- https://azure.github.io/kubelogin/ - the `convert-kubeconfig` step `kubeconfig` depends on
- `azure/az/README.md` for plugin wiring and the posh config sample
