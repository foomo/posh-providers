#### Hazards

Read-only. `detect` lists cluster resources whose apiVersions are deprecated or
removed; it never writes to the cluster, so it is safe to run unattended for
orientation. The only thing it mutates is the terminal.

Three flags are applied unconditionally by the provider and are not visible in
the usage block: `--ignore-removals`, `--ignore-deprecations` and
`--ignore-unavailable-replacements`. These suppress pluto's non-zero exit codes,
so **`detect` exits 0 even when it finds deprecated resources.** An agent must
read the output to decide whether anything was found; exit status carries no
signal here. `2>/dev/null` is also appended, so pluto's diagnostics - including
the reason for an empty result, such as a cluster it could not reach - are
discarded. An empty result is therefore indistinguishable from a failed
connection.

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

Everything after `detect` is forwarded to the real `pluto detect-all-in-cluster`,
as are any additional args and flags, so upstream flags not listed in the usage
block still work and inherit upstream behaviour.

#### Configuration

This provider reads no config key of its own. It consumes kubectl's, whose
`configPath` determines the selectable clusters and profiles - `kubectl` by
default, though that provider takes `CommandWithConfigKey`, so confirm the actual
key against the project's posh config.

Field shapes: [`kubernetes/kubectl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubectl/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same
schema is bundled under in the project's `posh.schema.json`.

#### Examples

```bash
# Cluster name comes before the verb
posh execute pluto my-cluster detect

# Scope to a profile and only report apiVersions already removed
posh execute pluto my-cluster detect --profile staging --only-show-removed

# Machine-readable output, since exit status never signals findings
posh execute pluto my-cluster detect --output json
```

#### References

- [pluto documentation](https://pluto.docs.fairwinds.com/)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/fairwindsops/pluto/README.md)
