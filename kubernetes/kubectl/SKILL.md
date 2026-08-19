#### Hazards

**The tree is not the surface.** Only the cluster name and `--profile` are
modelled; everything after the cluster is handed to the real `kubectl` binary
unchanged. So the entire kubectl CLI is reachable here - `delete`, `apply`,
`drain`, `exec` - and none of it appears in the usage block. Treat this as "all
of kubectl, pinned to one kubeconfig", not as the one argument shown.

**The cluster argument is the only thing standing between a local cluster and a
production one**, and it is the *first* word rather than a flag, so it is easy
to miss when reading a command back. There is no confirmation step and no
"current context" to fall back on: every invocation states its target
explicitly, which is safer than ambient `kubectl` but means a wrong name is
silently a different cluster rather than an error.

**An unknown cluster name is not rejected.** `Kubectl.Cluster(name)` *constructs*
a cluster rather than looking one up, so any string is accepted and turned into
a path. A typo therefore points `KUBECONFIG` at a file that does not exist, and
the failure surfaces from kubectl itself ("no configuration has been provided")
rather than from posh. Note for callers: a `Cluster(...)` return value is never
nil, so a nil-check on it can never fire - validate against the completion list
instead.

**Listing clusters rewrites file permissions.** Enumerating the config path -
which tab completion does - chmods every `*.yaml` in it to `0600` as a side
effect. That is a deliberate hardening of files holding credentials, but it
means simply completing a cluster name mutates the filesystem, and any
intentionally wider permission is silently reverted.

#### Behaviour

**This provider does not create kubeconfigs; it only selects among them.** A
cluster exists here because some other provider wrote
`<configPath>/<cluster>.yaml` - `gcloud`, `az`, `beam` and `k3d` all do - or
because a file was placed there by hand. If a cluster is missing, the fix is to
run whichever provider fetches its credentials, not anything in this command.

`--profile` does not reach kubectl. It selects a *subdirectory* of the config
path to read the kubeconfig from: with a profile, the file is
`<configPath>/<profile>/<cluster>.yaml`; without, `<configPath>/<cluster>.yaml`.
Profiles are therefore a way to keep several kubeconfigs for the same cluster
name side by side, and the suggestions come from the directories that exist in
the config path. Whatever you pass here must match what the provider that wrote
the kubeconfig was told to use.

Two different `KUBECONFIG` values are in play. Constructing the provider sets a
process-wide `KUBECONFIG=<configPath>/kubeconfig.yaml`, which every command in
the posh shell inherits - but nothing in this provider ever creates that file.
Each `kubectl <cluster>` invocation then overrides it with the selected
cluster's own file for that one call. So the ambient value is a fallback that is
usually a dangling path, and the per-command value is what actually applies.

Completion beyond the cluster name is not provided: flags and subcommands are
passed through verbatim without kubectl's own completion. Namespace and pod
listings *are* available to other providers via this package, cached for the
session and fetched with a 5-second timeout, which is why an unreachable cluster
shows an empty list rather than an error.

The cluster name `_` is a sentinel meaning "none" (`kubectl.None`), used by
providers that need a cluster-shaped value where no cluster applies.

#### Configuration

Config key `kubectl` by default, overridable via `CommandWithConfigKey` - note
that despite the `Command` prefix this is an option on `kubectl.New`, the
provider constructor, not on `NewCommand`. Confirm the key against the project's
own posh config.

Field shapes: [`kubernetes/kubectl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubectl/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

The single field `configPath` is the directory holding every kubeconfig. It is
created at startup if missing, and it is the shared contract between this
provider and every other one that writes a kubeconfig - so changing it moves
where those providers write, not just where this one reads.

An optional auth token provider can be wired in via
`CommandWithAuthTokenProvider`; when present, the namespace and pod lookups pass
`--token` to kubectl. It does not affect the passthrough command.

#### Examples

```bash
# Everything after the cluster goes straight to kubectl
posh execute kubectl my-cluster get pods -n my-namespace

# --profile picks a different kubeconfig for the same cluster name
posh execute kubectl my-cluster --profile azure get nodes
```

#### References

- [kubectl reference](https://kubernetes.io/docs/reference/kubectl/) - the real command surface behind the cluster argument
- [Provider README](https://github.com/foomo/posh-providers/blob/main/kubernetes/kubectl/README.md)
