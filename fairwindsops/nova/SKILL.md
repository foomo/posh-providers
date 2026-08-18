#### Hazards

Read-only. `find` lists Helm releases and container images in the cluster that
are behind their latest known version; it never writes to the cluster, so it is
safe to run unattended for orientation.

It is not offline, though. nova cross-checks what it finds against public Helm
repositories and container registries, so `find` makes outbound network calls
beyond the cluster and reports stale results - or none - when that egress is
blocked. `--show-errored-containers` is what surfaces those failures; without
it, a lookup that failed and a chart that is up to date look the same.

It is not offline in a second sense either: `--helm` and `--containers` select
what is scanned rather than only how it is displayed, and they compose. With
neither, nova defaults to Helm charts only, so a run that reports nothing about
images was never asked about them - a silent gap in coverage rather than a clean
result. Pass both to cover both in one invocation.

#### Behaviour

The cluster name is not free text. It is resolved from the `*.yaml` files present
in kubectl's `configPath`, so only clusters whose kubeconfig has already been
fetched are selectable - an unfetched cluster is not a typo but a missing file,
and the fix is whichever provider writes that kubeconfig (`gcloud`, `az`,
`teleport`, ...), not this command. `--profile` is resolved the same way, from the
*subdirectories* of `configPath`, and scopes the lookup to
`<configPath>/<profile>/<cluster>.yaml`. Both lists are cached for the shell
session, so a cluster fetched after the prompt started may not appear until the
cache is invalidated.

`find` and its flags are forwarded to the real `nova` binary, as are any
additional args and flags, so upstream flags not listed in the usage block still
work and inherit upstream behaviour. Unlike some providers here, stderr is not
suppressed - connection and lookup failures do reach the terminal.

#### Configuration

This provider reads no config key of its own. It consumes kubectl's, whose
`configPath` determines the selectable clusters and profiles - `kubectl` by
default, though that provider takes `CommandWithConfigKey`, so confirm the actual
key against the project's posh config.

Field shapes: [`kubernetes/kubectl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubectl/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same
schema is bundled under in the project's `posh.schema.json`.

`--config` points nova at *its own* config file, unrelated to the posh config;
`nova generate-config` writes a default one, and that verb is not exposed by this
tree.

#### Examples

```bash
# Cluster name comes before the verb
posh execute nova my-cluster find

# Both charts and images, only what is behind, with namespaces shown
posh execute nova my-cluster find --helm --containers --show-old --wide

# Machine-readable, scoped to a profile and a namespace
posh execute nova my-cluster find --profile staging --namespace ingress --format json
```

#### References

- [nova documentation](https://nova.docs.fairwinds.com/)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/fairwindsops/nova/README.md)
