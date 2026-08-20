#### Hazards

**This is a full-screen terminal UI and it does not exit on its own.** `k9s` takes over the terminal
and runs until a human quits it, so an agent must not invoke it — there is no output to parse and the
call blocks indefinitely. Use `kubectl` for anything an agent needs to read.

**k9s is read-write by default.** Unlike the sibling `ku`, there is no flag gating mutation: a human
at the keyboard can edit, scale and delete live resources, and the provider passes nothing that
restricts that. Whatever cluster and namespace the arguments resolve to is fully writable from the
moment the dashboard opens.

**The cluster argument is the only thing scoping the dashboard, so check it before running.** It is
validated — a name with no kubeconfig behind it is rejected as `invalid [cluster] argument`, and the
`--profile` flag is honoured while checking, so a cluster that exists only under a profile needs that
flag. What validation cannot catch is naming a *real* cluster you did not mean: `k9s prod` and
`k9s staging` differ by one word and both pass. Combined with the point above, that is a writable
dashboard on whichever cluster you named. See `kubectl` for the command that lists the configured
clusters.

**Omitting `[fleet]` is cluster-wide, not a default namespace.** With only a cluster given, no
`--namespace` is passed at all and the dashboard covers every namespace in the cluster. The usage
block shows `[fleet]` as optional, which reads as less scope rather than more.

#### Behaviour

**The command's arguments depend on how the project wired it.** `[fleet]` and `[squadron]` exist only
when the plugin passed `CommandWithSquadron`; without it the tree is `k9s <cluster>` and no namespace
is ever computed, so every invocation is cluster-wide. The generated usage block reflects one
project's wiring, so do not assume the other form is available.

**Namespace resolution has five outcomes, and two of them are probably not what was typed.** The
default `NamespaceFn` maps `(cluster, fleet, squadron)` as follows — verified by running the function
over each case:

| `[fleet]` | `[squadron]` | passed to `k9s` |
|---|---|---|
| omitted | anything | nothing — cluster-wide, and `[squadron]` is ignored entirely |
| `all` | omitted | `--all-namespaces` |
| `default` | omitted | `--namespace=default` |
| `default` | `shop` | `--namespace=shop` — the fleet name is dropped |
| `default` | `all` | `--all-namespaces` |
| `eu` | omitted | `--namespace=eu` |
| `eu` | `shop` | `--namespace=eu-shop` |
| `eu` | `all` | `--namespace=eu-all` — a literal namespace, **not** every namespace |

The last row is the trap: `all` is squadron's own sentinel for "every squadron" (`squadron.All`), and
it only widens the scope here when the fleet is omitted or `default`. Under any other fleet it is
concatenated like an ordinary name and k9s opens on a namespace that almost certainly does not exist.
Note also that `all` is never *suggested* for `[squadron]` — completion lists `squadron.List()`
verbatim, and unlike squadron's own tree this provider does not append the sentinel — so the only way
to reach `--all-namespaces` through the squadron argument is to type it.

A project can replace all of this with `CommandWithSquadronNamespaceFn`, so the table describes the
default and not a guarantee.

**Flags typed on the command are dropped; only what follows `--` reaches k9s.** `execute` forwards
`r.AdditionalArgs()` and `r.AdditionalFlags()` but never `r.Flags()`, so the only declared flag,
`--profile`, is consumed by posh to pick the kubeconfig and nothing else is passed through. To hand
k9s its own flags, put them after `--`.

**`--logoless` and `--splashless` are always passed.** The provider hardcodes both, so the dashboard
starts without the banner regardless of the user's k9s config.

#### Configuration

None. This provider reads no config key and ships no schema. The clusters it offers come from the
`kubectl` provider's config, and the fleets and squadrons from the `squadron` provider's — see those
two for the keys that actually control this command's arguments. `CommandWithSquadron` and
`CommandWithSquadronNamespaceFn` are wired in Go by the project's plugin, not in YAML.

#### Examples

```bash
# every namespace in the cluster
posh execute k9s prod

# a specific fleet, when the project wired a squadron
posh execute k9s prod eu

# fleet plus squadron -> namespace eu-shop
posh execute k9s prod eu shop

# every namespace, explicitly
posh execute k9s prod all

# pass flags to k9s itself after -- (nothing typed before -- reaches it)
posh execute k9s prod eu -- --command pods

# pick a non-default kubeconfig profile
posh execute k9s prod --profile azure
```

#### References

- [k9s docs](https://k9scli.io/) — the dashboard itself, its keybindings and its own config
- [provider README](README.md)
