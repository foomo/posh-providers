#### Hazards

**This is a full-screen terminal UI and it does not exit on its own.** It takes over the
terminal until a human quits it, so an agent must not invoke it - there is no output to parse
and the call blocks indefinitely. Use `kubectl` for anything an agent needs to read.

**`--edit` makes the dashboard read-write against the selected cluster.** Without it the
dashboard is read-only. With it, a human can edit and delete live resources, so the flag both
requires approval and widens what a wrong cluster argument can damage.

**The cluster argument is the only thing scoping the session.** It is validated - a name with
no kubeconfig is rejected as `invalid [cluster] argument`, and `--profile` is honoured while
checking, so a cluster existing only under a profile needs that flag. What validation cannot
catch is naming a *real* cluster you did not mean.

**Omitting `[fleet]` is cluster-wide, not a default namespace.** With only a cluster given no
`--namespace` is passed and the dashboard covers every namespace - including whatever
`--edit` then permits. The usage block shows `[fleet]` as optional, which reads as less scope
rather than more.

#### Behaviour

**The arguments depend on how the project wired it.** `[fleet]` and `[squadron]` exist only
when the plugin passed `CommandWithSquadron`; without it the tree is one cluster argument and
no namespace is computed.

**Namespace resolution drops information in two cases.** The default mapping: nothing when
the fleet is omitted; `<squadron>` alone when the fleet is `default`, dropping the fleet
name; `<fleet>` when the squadron is omitted; `<fleet>-<squadron>` otherwise. A fleet of
`all` alone means cluster-wide, but with a squadron it concatenates, giving `all-shop`.
`CommandWithSquadronNamespaceFn` replaces the mapping.

**`[squadron]` completion ignores the cluster and fleet already chosen**, offering every
squadron in the project including ones not deployed to the selected fleet. `[fleet]`
completion does respect the cluster. Neither is validated.

`--profile` is consumed by posh; everything after `--` reaches the binary verbatim.

#### Configuration

None of its own - no config key, no schema. The cluster list and `configPath` come from
`kubernetes/kubectl`, the fleet list from `foomo/squadron` (`clusters[].fleets`).

#### Examples

```bash
# read-only, whole cluster
posh execute {{cmd}} prod

# one squadron within a fleet -> --namespace=eu-shop
posh execute {{cmd}} prod eu shop --profile myprofile

# read-write - a human must drive this
posh execute {{cmd}} prod eu --edit
```

#### References

- [bjarneo/ku](https://github.com/bjarneo/ku)
- [provider README](https://github.com/foomo/posh-providers/blob/main/bjarneo/ku/README.md)
