#### Hazards

**The cluster name is the first word, and it decides what every subcommand
mutates.** `helm <cluster> uninstall <release>` and the same line against a
production cluster differ by one argument, with no confirmation step from posh.
The destructive verbs here are `uninstall` (removes a release and its
resources), `upgrade` and `install` (apply changes to the cluster), and
`rollback` (moves a release to another revision). `plugin` and `repo` mutate the
local helm installation rather than the cluster. Everything else - `list`,
`status`, `get`, `history`, `show`, `search`, `env`, `version`, `lint`,
`template`, `diff` - is read-only or local.

**`--dry-run` is only as safe as the subcommand honours it.** It is offered on
every leaf here, including ones upstream helm does not support it on, because
the flag set is shared. On those it is forwarded and rejected by helm rather
than silently ignored, but do not treat its presence in the usage block as proof
a given verb can be simulated.

**The cluster argument is the only thing scoping every verb, so check it before
running.** It is validated against the kubeconfig that will actually be used -
the `--profile` flag is honoured, so a cluster that exists only under a profile
is accepted with that flag and rejected without it. What validation cannot catch
is naming a *real* cluster you did not mean, which for `uninstall` and `rollback`
is the case that matters.

#### Behaviour

The tree is a hand-maintained mirror of helm's subcommands, not a passthrough
root, so it can drift from the installed binary: a subcommand helm gained since
this list was written is not reachable, and every leaf dispatches to the same
`Execute` that simply forwards its own name. Arguments the tree does not model -
release names, chart paths, `-f values.yaml` - still work; they are passed
through, and additional args and flags after a `--` separator are forwarded too.

Only flags you actually type are forwarded. The shared flag set is broad
(`--namespace`, `--atomic`, `--wait`, `--create-namespace`, ...) but execution
sends `fs.Visited()`, so unset flags are not passed as defaults and helm's own
defaults apply.

This provider does not create kubeconfigs; it selects among the ones
`kubernetes/kubectl` manages. A cluster is selectable because some other
provider - `gcloud`, `az`, `beam`, `k3d` - wrote `<configPath>/<cluster>.yaml`.
If the cluster you want is missing, run that provider, not anything here.
`--profile` picks a subdirectory of the same config path rather than reaching
helm.

The kubeconfig is passed as a `KUBECONFIG` environment variable scoped to the
one invocation, so it does not disturb the ambient value other commands in the
shell see.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type and
no `config.schema.json`, so nothing to set in the project's posh config and
nothing for it in `posh.schema.json`. Everything it needs comes from the
`kubernetes/kubectl` provider it is constructed with, so the relevant config is
kubectl's `configPath`: see
[`kubernetes/kubectl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubectl/config.schema.json).

The command can be renamed with `CommandWithName`, and the helm binary
invocation swapped with `CommandWithExecHelm` - both options on `NewCommand`.

#### Examples

```bash
# Read-only
posh execute helm my-cluster list --namespace my-namespace

# Mutating - the cluster argument is the only thing scoping this
posh execute helm my-cluster upgrade my-release ./chart --atomic --wait

# Same cluster, kubeconfig read from a profile subdirectory
posh execute helm my-cluster --profile azure list
```

#### References

- [Helm documentation](https://helm.sh/docs/) - the real behaviour of each subcommand
- [`kubernetes/kubectl`](https://github.com/foomo/posh-providers/blob/main/kubernetes/kubectl/README.md) - owns the kubeconfigs this command selects from
- [Provider README](https://github.com/foomo/posh-providers/blob/main/helm/helm/README.md)
