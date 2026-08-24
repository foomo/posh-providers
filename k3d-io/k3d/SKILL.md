#### Hazards

**`down` is a delete, not a shutdown.** Despite reading like the opposite of `up`, it runs the
binary's `cluster delete` and removes the kubeconfig from disk - the cluster is gone, not stopped.
Use `pause` and `resume` to stop and start. Requires approval.

The config holds exactly **one** `registry:`, shared by every cluster and created by the first `up`,
so `down` removes it only once the last configured cluster is gone.

**`install` runs `helm upgrade --install --force --dependency-update`.** `--force` replaces resources
rather than patching, so an immutable field means delete-and-recreate: for a Deployment the pods go
away and come back. `--dependency-update` re-fetches chart dependencies from the network every run.

**`install` and `uninstall` accept any cluster name.** The other five verbs reject a name not under
`clusters:`; these two pass it straight to the shared kubeconfig directory every other provider writes
into, so any cluster with a kubeconfig there is a valid helm target and a name colliding with a real
cluster lands the chart on it.

#### Behaviour

**`up` is idempotent and silent about it.** If the cluster exists it logs `cluster already exists` and
returns success without reconciling image, port, args or traefik against the config. Editing a
cluster's config and re-running `up` therefore appears to work and changes nothing; `down` then `up`
applies it.

**The real cluster name is not necessarily the argument.** Every verb takes the config *key* and acts
on `alias:` if one is set: with `alias: foomo`, `{{cmd}} up local` creates a cluster the binary lists
as `foomo`. The kubeconfig is still written under the **argument** name.

**Traefik is disabled unless `enableTraefikRouter` is true**, the inverse of k3s's default.

**Chart names come from a directory listing, not from the cluster.** The chart argument completes to
subdirectories of `charts.path` regardless of what is installed - including for `uninstall`. The
namespace is `<charts.prefix><chart>`, created if absent. A
`<charts.path>/<chart>/values.override.yaml` is passed as `--values` silently, with no way to skip it.

**The registry's real hostname carries a `k3d-` prefix** added to `registry.name`, so images must be
pushed to `k3d-<name>:<port>`.

#### Configuration

The `k3d` key holds `charts`, the one shared `registry`, and `clusters` keyed by the argument name.

Field shapes: [`k3d-io/k3d/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/k3d-io/k3d/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} up local
posh execute {{cmd}} install local base
posh execute {{cmd}} up local -- --servers 3
```

#### References

- [k3d docs](https://k3d.io/) - the binary every verb wraps
- [provider README](https://github.com/foomo/posh-providers/blob/main/k3d-io/k3d/README.md)
