#### Hazards

**The cluster name is the first word, and it decides what every subcommand
mutates** - a local and a production invocation differ by one argument, with no
confirmation step from posh. The destructive verbs are `uninstall` (removes a
release and its resources), `upgrade` and `install`, and `rollback`. `plugin` and
`repo` mutate the local installation rather than the cluster. Everything else -
`list`, `status`, `get`, `history`, `show`, `search`, `env`, `version`, `lint`,
`template`, `diff` - is read-only or local.

**`--dry-run` is only as safe as the subcommand honours it.** The flag set is
shared by every leaf, so it is offered on ones upstream does not support it on -
where it is forwarded and rejected. Do not treat its presence in the usage block
as proof a verb can be simulated.

**The cluster argument is validated** against the kubeconfig that will actually be
used - `--profile` is honoured, so a cluster existing only under a profile is
accepted with that flag and rejected without it. What validation cannot catch is
naming a *real* cluster you did not mean, which for `uninstall` and `rollback` is
the case that matters.

#### Behaviour

The tree is a hand-maintained mirror of the binary's subcommands, not a
passthrough root, so a subcommand upstream gained since this list was written is
not reachable. Arguments the tree does not model - release names, chart paths,
`-f values.yaml` - still work, as do args and flags after a `--`.

Only flags you actually type are forwarded, so unset ones are not sent as
defaults and the binary's own defaults apply.

This provider does not create kubeconfigs; it selects among the ones
`kubernetes/kubectl` manages. A cluster is selectable because some other provider
- `gcloud`, `az`, `beam`, `k3d` - wrote `<configPath>/<cluster>.yaml`; if the one
you want is missing, run that provider. `--profile` picks a subdirectory of the
same path rather than reaching the binary.

#### Configuration

**No config key of its own** - no `Config` type, no `config.schema.json`.
Everything it needs comes from the `kubernetes/kubectl` provider it is constructed
with, so the relevant config is kubectl's `configPath`:
[`kubernetes/kubectl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubectl/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} my-cluster list --namespace my-namespace

# Mutating - the cluster argument is the only thing scoping this
posh execute {{cmd}} my-cluster upgrade my-release ./chart --atomic --wait
```

#### References

- [Helm documentation](https://helm.sh/docs/) - what each subcommand really does
- [`kubernetes/kubectl`](https://github.com/foomo/posh-providers/blob/main/kubernetes/kubectl/README.md) - owns the kubeconfigs
