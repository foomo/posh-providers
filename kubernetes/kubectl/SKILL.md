#### Hazards

**The tree is not the surface.** Only the cluster name and `--profile` are
modelled; everything after the cluster goes to the real `kubectl` binary
unchanged. So the entire upstream CLI is reachable here - `delete`, `apply`,
`drain`, `exec` - and none of it appears in the usage block. Treat this as the
whole binary pinned to one kubeconfig.

**The cluster is the *first* word rather than a flag**, so it is easy to miss
when reading a command back, and it is the only thing between a local cluster and
a production one. No confirmation step, no current-context fallback.

**An unknown cluster name is not rejected.** `Kubectl.Cluster(name)`
*constructs* a cluster rather than looking one up, so any string becomes a path.
A typo points `KUBECONFIG` at a file that does not exist and the failure surfaces
from the binary ("no configuration has been provided"), not from posh. That
return is never nil, so a nil-check can never fire - validate against the
completion list.

**Listing clusters rewrites file permissions.** Enumerating the config path -
which tab completion does - chmods every `*.yaml` in it to `0600`, once per
session. Any intentionally wider permission is silently reverted.

#### Behaviour

**This provider does not create kubeconfigs; it only selects among them.** A
cluster exists here because another provider wrote
`<configPath>/<cluster>.yaml` - `gcloud`, `az`, `beam` and `k3d` all do. If one
is missing, run whichever provider fetches its credentials.

`--profile` does not reach the binary. It selects a *subdirectory* of the config
path, so the file becomes `<configPath>/<profile>/<cluster>.yaml`. It must match
what the provider that wrote the kubeconfig was told to use.

Two `KUBECONFIG` values are in play. Construction sets a process-wide
`KUBECONFIG=<configPath>/kubeconfig.yaml` that every command in the shell
inherits - but nothing here ever creates that file. Each invocation then
overrides it with the selected cluster's own file, so the ambient value is a
usually-dangling fallback.

#### Configuration

The `kubectl` key's single field `configPath` is the directory holding every
kubeconfig. It is created at startup if missing, and it is the shared contract
between this provider and every other one that writes a kubeconfig - changing it
moves where those providers write, not just where this one reads.

Field shapes: [`kubernetes/kubectl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubectl/config.schema.json).

#### References

- [kubectl reference](https://kubernetes.io/docs/reference/kubectl/) - the real surface behind the cluster argument
- [Provider README](https://github.com/foomo/posh-providers/blob/main/kubernetes/kubectl/README.md)
