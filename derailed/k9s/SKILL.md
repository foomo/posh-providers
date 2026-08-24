#### Hazards

**This is a full-screen terminal UI and it does not exit on its own.** It takes over the
terminal until a human quits it, so an agent must not invoke it - nothing to parse and the
call blocks indefinitely. Use `kubectl` for anything an agent needs to read.

**It is read-write from the moment it opens.** Unlike the sibling `ku`, there is no flag
gating mutation: a human can edit, scale and delete live resources, and the provider passes
nothing that restricts that.

**The cluster argument is the only thing scoping the dashboard.** It is validated - a name
with no kubeconfig is rejected as `invalid [cluster] argument`, and `--profile` is honoured
while checking, so a cluster existing only under a profile needs that flag. What validation
cannot catch is naming a *real* cluster you did not mean, which with the point above is a
writable dashboard on it.

**Omitting `[fleet]` is cluster-wide, not a default namespace.** With only a cluster given no
`--namespace` is passed and every namespace is covered. The usage block shows `[fleet]` as
optional, which reads as less scope rather than more.

#### Behaviour

**The arguments depend on how the project wired it.** `[fleet]` and `[squadron]` exist only
when the plugin passed `CommandWithSquadron`; without it every invocation is cluster-wide.

**Namespace resolution has one real trap.** The default mapping: nothing when the fleet is
omitted, and `[squadron]` is then ignored entirely; `<squadron>` alone when the fleet is
`default`, dropping the fleet name; `<fleet>` when the squadron is omitted;
`<fleet>-<squadron>` otherwise. `--all-namespaces` is passed only when that result is exactly
`all`, so `eu all` yields `--namespace=eu-all` - a literal namespace that almost certainly
does not exist. `CommandWithSquadronNamespaceFn` replaces the mapping.

**Flags typed on the command are dropped; only what follows `--` reaches the binary**, which
is always given `--logoless` and `--splashless`. `--profile` is consumed by posh.

#### Configuration

None. No config key, no schema. The clusters come from the `kubectl` provider's config, the
fleets and squadrons from `foomo/squadron`'s; both `CommandWith...` options are wired in Go
by the project's plugin, not in YAML.

#### Examples

```bash
# every namespace in the cluster
posh execute {{cmd}} prod

# fleet plus squadron -> namespace eu-shop; nothing before -- reaches the binary
posh execute {{cmd}} prod eu shop --profile azure -- --command pods
```

#### References

- [k9s docs](https://k9scli.io/) - keybindings and its own config
- [provider README](https://github.com/foomo/posh-providers/blob/main/derailed/k9s/README.md)
