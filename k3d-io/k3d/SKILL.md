#### Hazards

**`down` deletes the shared registry, even while other clusters are still using it.** The config holds
exactly **one** `registry:` for all clusters, and `down` deletes it whenever it exists — the code
carries the author's own `// TODO check if empty` at that line. So with two clusters up, tearing one
down removes the registry the other is pulling from, and that cluster keeps running with a broken
image source. Bring the registry back with `up` on any cluster. Verified by tracing the config shape
against the teardown order.

**`down` is a delete, not a shutdown.** Despite reading like the opposite of `up`, it runs
`k3d cluster delete`, removes the kubeconfig from disk and then deletes the registry — everything in
the cluster is gone, not stopped. The verbs for stopping and starting are `pause` and `resume`.
Requires approval.

**`install` runs `helm upgrade --install --force`.** `--force` makes helm replace resources rather than
patch them, so immutable fields are handled by delete-and-recreate — which for a Deployment or a
StatefulSet means the pods go away and come back. It also passes `--dependency-update`, so chart
dependencies are re-fetched from the network on every run. Fine against a local throwaway cluster,
which is what this provider targets; worth knowing before pointing it anywhere else.

**Nothing stops these verbs from acting on a non-k3d cluster.** The name argument is checked against
`clusters:` in config, so an unknown name is rejected — but the `KUBECONFIG` the helm verbs run under
comes from `kubectl.Cluster(<name>).Env("")`, the same shared kubeconfig directory every other
provider writes to. If a configured k3d cluster name collides with a real cluster's kubeconfig, the
chart lands there instead. See `kubectl` for what that directory holds.

#### Behaviour

**`up` is idempotent and silent about it.** If the cluster already exists it logs `cluster already
exists` and returns success without touching anything — it does not reconcile image, port, args or
traefik settings against the current config. Changing a cluster's config and re-running `up` therefore
appears to work and changes nothing; `down` then `up` is what applies it.

**`up` creates the registry only when it is missing**, so the first cluster brought up creates it and
later ones reuse it. Combined with the teardown hazard above, registry lifetime is "from the first
`up` to the first `down`", not per-cluster.

**The k3d cluster name is not necessarily the argument.** Every verb takes the config *key* as its
argument, then acts on `alias:` if the cluster sets one, falling back to the key. So `k3d up local`
with `alias: foomo` creates a cluster k3d knows as `foomo`, and `k3d cluster list` will not show
`local`. The kubeconfig, however, is written under the **argument** name, not the alias.

**Traefik is disabled unless `enableTraefikRouter` is true.** The provider passes
`--k3s-arg "--disable=traefik@server:0"` by default, which is the inverse of k3s's own default. The
port mapping is always `<port>:443@loadbalancer`, so the configured `port:` reaches the cluster's
HTTPS listener.

**`--agents 1` is hardcoded**, as is `K3D_FIX_DNS=1`, which the provider exports process-wide at
construction. Neither can be configured; extra `k3d cluster create` arguments can be added through the
cluster's `args:` list or after `--`.

**Chart names come from directory listing, not from the cluster.** The chart argument completes to the
subdirectory names under `charts.path`, so it offers charts that exist on disk regardless of what is
installed — including for `uninstall`, where a chart never installed completes just as readily. The
namespace is `<charts.prefix><chart>` and `install` creates it if absent.

**`install` picks up an override file automatically.** If `<charts.path>/<chart>/values.override.yaml`
exists it is passed as `--values`; nothing reports that it was applied, and there is no flag to skip
it.

**`kubeconfig` clears the provider cache after writing**, so cluster and chart listings are re-read on
the next completion. The other verbs do not, and `k3d`'s own cluster/registry lookups shell out fresh
each time rather than caching.

**The registry's real hostname carries a `k3d-` prefix.** `registry.name` is what is passed to k3d,
which names the actual container and in-cluster host `k3d-<name>`. Images must be pushed to
`k3d-<name>:<port>`, not `<name>:<port>`.

#### Configuration

Read from the `k3d` key by default. The override is `WithConfigKey` — an option on **`k3d.New`**, not
on `NewCommand`, which takes `CommandWithName` to rename the command. The provider is built in two
pieces: `k3d.New(l)` for the client and `NewCommand(l, k3d, cache, kubectl)` for the command. It also
ships a `Checker` for the posh prompt.

Note the README's plugin snippet calls `k3d.NewCommand(l, inst.k3d, inst.kubectl)` — three arguments,
where the real signature takes four (`cache` sits between) and also returns an error. As written it
does not compile.

- [config.schema.json](https://raw.githubusercontent.com/foomo/posh-providers/main/k3d-io/k3d/config.schema.json)

#### Examples

```bash
# create the registry (first time) and the cluster
posh execute k3d up local

# write its kubeconfig into the shared kubectl config directory
posh execute k3d kubeconfig local

# install a chart from charts.path into <prefix><chart>
posh execute k3d install local base

# stop and start without destroying anything
posh execute k3d pause local
posh execute k3d resume local

# remove the chart
posh execute k3d uninstall local base

# delete the cluster, its kubeconfig AND the shared registry
posh execute k3d down local

# extra k3d flags after --
posh execute k3d up local -- --servers 3
```

#### References

- [k3d docs](https://k3d.io/) — the binary every verb wraps
- [provider README](README.md) — note its plugin snippet omits the `cache` argument and does not compile
