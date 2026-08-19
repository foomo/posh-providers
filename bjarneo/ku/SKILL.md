#### Hazards

**This is a full-screen terminal UI and it does not exit on its own.** `ku` takes over the terminal
and runs until a human quits it, so an agent must not invoke it — there is no output to parse and the
call blocks indefinitely. Use `kubectl` for anything an agent needs to read.

**`--edit` makes the dashboard read-write against the selected cluster.** Without it `ku` starts
read-only. With it, a human at the keyboard can edit and delete live resources, so the flag both
requires approval and widens what a mistyped cluster argument can damage. There is no dry-run.

**A cluster argument is never validated.** The provider builds `KUBECONFIG` from the name it is given
(`kubectl.Cluster(<name>).Env(<profile>)`) without checking that the file exists — `kubectl`'s own
`ConfigExists` helper is not called here. A typo therefore points `KUBECONFIG` at a nonexistent path
rather than failing, and `ku` starts against whatever the ambient kube context resolves to, which may
be a different cluster than the one named. Confirm the cluster before trusting the argument, and see
`kubectl` for the command that lists the configured ones.

**Omitting `[fleet]` is cluster-wide, not a default namespace.** With only a cluster given, no
`--namespace` is passed at all and the dashboard covers every namespace in the cluster — including
whatever `--edit` then permits. The usage block shows `[fleet]` as optional, which reads as less
scope rather than more. See Behaviour for how the namespace is actually computed.

#### Behaviour

**The command's arguments depend on how the project wired it.** `[fleet]` and `[squadron]` exist only
when the plugin passed `CommandWithSquadron`; without it the tree is `ku <cluster>` and no namespace
is ever computed. The generated usage block reflects one project's wiring, so do not assume the other
form is available.

**Namespace resolution has four branches and two of them drop information.** The default
`NamespaceFn` maps `(cluster, fleet, squadron)` as follows — verified by running the function over
each case:

| `[fleet]` | `[squadron]` | passed to `ku` |
|---|---|---|
| omitted | anything | nothing — cluster-wide |
| `default` | omitted | `--namespace=default` |
| `default` | `shop` | `--namespace=shop` — the fleet name is dropped |
| `eu` | omitted | `--namespace=eu` |
| `eu` | `shop` | `--namespace=eu-shop` |
| `all` | omitted | nothing — cluster-wide |
| `all` | `shop` | `--namespace=all-shop` — probably not a real namespace |

So `all` is special-cased to mean cluster-wide, but only on its own: combined with a squadron it is
concatenated like any other fleet. A project can replace the whole mapping with
`CommandWithSquadronNamespaceFn`, so treat the table as this provider's default rather than as the
project's behaviour.

**`[squadron]` completion ignores the cluster and fleet already chosen.** It calls `squadron.List()`,
which takes no arguments, so it offers every squadron in the project — including ones not deployed to
the selected fleet. `[fleet]` completion does respect the cluster (it reads that cluster's configured
`fleets`). Neither is validated on execute.

**`--profile` selects a kubeconfig subdirectory and only appears once a cluster is typed.** It is an
internal flag, so it is consumed by posh rather than forwarded, and it changes the config path from
`<configPath>/<cluster>.yaml` to `<configPath>/<profile>/<cluster>.yaml`. Its suggestions are the
directories under `kubectl`'s `configPath`, which is why they are the same for every cluster.

**Everything after `--` reaches the `ku` binary**, including flags: the provider forwards both
`AdditionalArgs()` and `AdditionalFlags()`, so upstream options this tree does not model are still
usable. Note the two flags it does model (`--edit`, `--dev`) are re-emitted as `ku`'s own
`--edit`/`--dev`, and `--namespace=` is always passed as a single token.

#### Configuration

None of its own — this provider reads no config key and ships no schema. It borrows two:
the cluster list and `configPath` come from `kubernetes/kubectl`, and the fleet list from
`foomo/squadron` (its `clusters[].fleets`). Read those providers' schemas for the field shapes, and
the project's own config file for the values.

#### Examples

```bash
# read-only dashboard, whole cluster
posh execute ku prod

# scoped to a fleet's namespace
posh execute ku prod eu

# scoped to one squadron within a fleet -> --namespace=eu-shop
posh execute ku prod eu shop

# non-default kubeconfig profile
posh execute ku prod eu --profile myprofile

# read-write mode - a human must drive this, and it can delete live resources
posh execute ku prod eu --edit
```

Anything after `--` is forwarded verbatim to the `ku` binary; check its own `--help` for what it
accepts, since this tree models only `--edit` and `--dev`.

#### References

- [bjarneo/ku](https://github.com/bjarneo/ku) — the upstream TUI
- [provider README](README.md)
