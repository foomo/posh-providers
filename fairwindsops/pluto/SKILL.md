#### Hazards

Read-only. `detect` lists cluster resources whose apiVersions are deprecated or
removed; it never writes to the cluster, so it is safe to run unattended.

Three flags are applied unconditionally by the provider and appear nowhere in the
usage block: `--ignore-removals`, `--ignore-deprecations` and
`--ignore-unavailable-replacements`. They suppress the upstream non-zero exit
codes, so **`detect` exits 0 even when it finds deprecated resources.** An agent
must read the output to decide whether anything was found; exit status carries no
signal here.

`2>/dev/null` is also appended, so the upstream diagnostics - including the reason
for an empty result, such as a cluster it could not reach - are discarded. An
empty result is therefore indistinguishable from a failed connection. Confirm
connectivity another way before reporting a cluster clean.

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

The verb maps to the upstream `detect-all-in-cluster` subcommand, and additional
args and flags are forwarded, so upstream flags absent from the usage block still
work.

#### Configuration

This provider reads no config key of its own. It consumes kubectl's, whose
`configPath` determines the selectable clusters and profiles.

Field shapes: [`kubernetes/kubectl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubectl/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} my-cluster detect
posh execute {{cmd}} my-cluster detect --profile staging --only-show-removed
posh execute {{cmd}} my-cluster detect --output json
```

#### References

- [pluto documentation](https://pluto.docs.fairwinds.com/)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/fairwindsops/pluto/README.md)
