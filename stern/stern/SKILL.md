#### Hazards

Every verb here is **read-only**: stern tails pod logs and this provider adds no verb
that writes to a cluster. Nothing needs approval on the grounds of mutating state.

What it does need is a caveat about termination and about output. **`stern` streams and
does not exit** — it follows logs until interrupted, so an agent invoking any of these
three verbs will block indefinitely rather than getting a result. Bound it explicitly:
`--tail <n>` limits the backlog but does **not** stop the follow, so pass `--` and
stern's own `--no-follow` when you want a finite read. Treat an unbounded invocation as
needing a human at the terminal.

Log output is not filtered for secrets. Tailing a workload that logs tokens, connection
strings or request bodies puts them in the transcript, and `--all-namespaces` (available
on `query` and `raw`) widens that to the whole cluster. Prefer the narrowest target you
can name, and prefer `squadron` or a named `query` over `raw '.*'`.

The cluster is the **first** argument, so a staging and a production tail differ by one
word. It is a dynamic tree node rather than a plain argument, so only names with an
existing kubeconfig match: a typo is rejected as `invalid command` by the tree itself,
before anything runs. That protects against mistyping but not against naming the wrong
*real* cluster, which is the case to check.

The root has no `Execute`, so this is not a passthrough: only `query`, `raw` and
`squadron` exist. Everything else stern offers is reachable only by putting it after
`--`.

#### Behaviour

**`query` walks a recursive tree and concatenates as it descends.** Each argument
selects one level of the nested `queries` config, and the `query` arguments of *every*
level walked are appended in order — so `query all errors` runs stern with `all`'s
arguments followed by `errors`'. A nested entry refines its parent rather than replacing
it, which is why a child usually holds only `--include` or `--container` rather than a
pod regex.

Two consequences worth knowing. An **unmatched name after a valid one is silently
dropped**: `query all bogus` resolves to just `all` and tails it, with no warning — only
a bad *first* name fails, with `query not found`. And completion reflects the walk, so
after a full path is given the suggestion list is empty rather than repeating the
siblings. Verified by running the lookup over a nested config.

`raw` passes its argument straight to stern as the pod query, so it is a **regex**, not
a name — `raw '.*'` tails everything in the namespace, and `--all-namespaces` widens it
to the cluster.

`squadron` derives both halves of its target: pods are matched as `<squadron>-<unit>`
and the namespace comes from a function of cluster, fleet and squadron. The default is
the bare squadron name when the fleet is `default`, and `<fleet>-<squadron>` otherwise —
`CommandWithNamespaceFn` replaces that derivation, so in a project passing it the
namespace may follow a different rule entirely. Note `squadron` declares no
`--namespace` or `--all-namespaces` flag precisely because it computes the namespace
itself.

Only flags actually typed are forwarded (`fs.Visited().Args()`), so nothing
posh-internal leaks to the binary — `--profile` is consumed by posh to pick the
kubeconfig. Flags declared here mirror stern's own names and are passed through
verbatim. Note three are declared with the literal default `"default"`: harmless for
`--output`, where `default` really is one of stern's values, but meaningless for
`--since` (a duration) and `--template` (a Go template). Because only *typed* flags are
forwarded those bogus defaults are never sent — but the completion list offers them, so
`--since default` completes and then fails inside stern.

Namespace completion on `query` and `raw` is a live `kubectl get namespaces` against the
selected cluster, with no profile applied, so it can be empty or slow when the cluster
is unreachable. Fleet, squadron and unit completions come from the `squadron` provider.

#### Configuration

Read under the `stern` key by default; `CommandWithConfigKey` renames it and
`CommandWithName` renames the command, both `CommandOption`s on `NewCommand`. Confirm
against the project's own config file. Field shapes:
[`stern/stern/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/stern/stern/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`queries` is the whole vocabulary the `query` verb accepts — read it to learn which
names exist at each level and what arguments each contributes. Where the kubeconfig
comes from is governed by the `kubectl` provider's own `configPath`.

#### Examples

```bash
# Tail a named query. Streams until interrupted.
posh execute stern prod query all

# Descend the query tree; `errors` refines `all` rather than replacing it.
posh execute stern prod query all errors

# Finite read — bound the backlog and stop following.
posh execute stern prod query all --tail 100 -- --no-follow

# A squadron unit; namespace is derived from fleet and squadron.
posh execute stern prod squadron default myapp myunit

# Raw pod regex. Prefer a narrower target than '.*'.
posh execute stern prod raw 'api-.*'
```

#### References

- [stern usage](https://github.com/stern/stern#usage)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/stern/stern/README.md) — its config sample is a good illustration of the nested `queries` shape; note the `Plugin` struct in its snippet has no `squadron` field despite using one
