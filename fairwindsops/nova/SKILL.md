#### Hazards

Read-only. `find` lists Helm releases and container images in the cluster that
are behind their latest known version; it never writes to the cluster, so it is
safe to run unattended for orientation.

It is not offline. The scan cross-checks findings against public Helm
repositories and container registries, so `find` makes outbound calls beyond the
cluster and reports stale results - or none - when that egress is blocked.
`--show-errored-containers` is what surfaces those failures; without it a lookup
that failed and a chart that is up to date look identical.

`--helm` and `--containers` select **what is scanned**, not only how it is
displayed, and they compose. With neither, the default is Helm charts only, so a
run reporting nothing about images was never asked about them - a silent gap in
coverage rather than a clean result. Pass both to cover both in one invocation.

#### Behaviour

The cluster name is not free text. It resolves from the `*.yaml` files present in
kubectl's `configPath`, so only clusters whose kubeconfig has already been
fetched are selectable - an unfetched cluster is a missing file, not a typo, and
the fix is whichever provider writes that kubeconfig (`gcloud`, `az`,
`teleport`, ...), not this command. `--profile` resolves the same way from the
*subdirectories* of `configPath`, scoping the lookup to
`<configPath>/<profile>/<cluster>.yaml`. Both lists are cached for the shell
session, so a cluster fetched after the prompt started may not appear until the
cache is invalidated.

The verb and its flags are forwarded to the upstream binary, as are additional
args and flags, so upstream flags absent from the usage block still work. Unlike
the sibling `pluto` command, stderr is not suppressed - connection and lookup
failures do reach the terminal.

#### Configuration

This provider reads no config key of its own. It consumes kubectl's, whose
`configPath` determines the selectable clusters and profiles.

Field shapes: [`kubernetes/kubectl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubectl/config.schema.json).

`--config` points the upstream binary at *its own* config file, unrelated to the
posh config. The `generate-config` verb that writes a default one is not exposed
by this tree.

#### Examples

```bash
posh execute {{cmd}} my-cluster find --helm --containers --show-old --wide
posh execute {{cmd}} my-cluster find --profile staging --namespace ingress --format json
```

#### References

- [nova documentation](https://nova.docs.fairwinds.com/)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/fairwindsops/nova/README.md)
