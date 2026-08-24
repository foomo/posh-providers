#### Hazards

Every verb here is **read-only**: it tails pod logs and adds nothing that writes to a
cluster, so nothing needs approval for mutating state.

**The binary streams and does not exit.** It follows logs until interrupted, so an agent
invoking any of these verbs blocks indefinitely rather than getting a result. `--tail <n>`
limits the backlog but does **not** stop the follow; pass `--` and the binary's own
`--no-follow` for a finite read.

Output is not filtered for secrets: a workload logging tokens or request bodies puts them in
the transcript, and `--all-namespaces` widens that to the whole cluster. Prefer a named
`query` or `squadron` over `raw '.*'`.

The cluster is the **first** argument, so a staging and a production tail differ by one
word. Only names with an existing kubeconfig match, so a typo is rejected by the tree - but
that does not catch naming the wrong *real* cluster.

The root has no `Execute`, so this is not a passthrough: anything else is reachable only
after `--`.

#### Behaviour

**`query` walks a recursive tree and concatenates as it descends.** Each argument selects
one level of the nested `queries` config and every level walked contributes its own `query`
arguments in order, so a nested entry refines its parent rather than replacing it. An
**unmatched name after a valid one is silently dropped**: `query all bogus` resolves to just
`all` and tails it with no warning; only a bad *first* name fails, with `query not found`.

`raw` passes its argument through as the pod query, so it is a **regex**, not a name.

`squadron` derives both halves of its target: pods are matched as `<squadron>-<unit>`, and
the namespace is the bare squadron name when the fleet is `default`, `<fleet>-<squadron>`
otherwise - replaceable with `CommandWithNamespaceFn`. It declares no `--namespace` because
it computes one.

`--profile` is consumed by posh to pick the kubeconfig rather than passed on, and only flags
actually typed are forwarded.

#### Configuration

The `stern` key's `queries` map is the whole vocabulary the `query` verb accepts - read it
for the names that exist at each level.

Field shapes: [`stern/stern/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/stern/stern/config.schema.json).

#### Examples

```bash
# Streams until interrupted; `errors` refines `all`
posh execute {{cmd}} prod query all errors

# Finite read - bound the backlog and stop following
posh execute {{cmd}} prod query all --tail 100 -- --no-follow
```

#### References

- [stern usage](https://github.com/stern/stern#usage)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/stern/stern/README.md)
