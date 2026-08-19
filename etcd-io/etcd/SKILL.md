#### Hazards

`edit` writes directly to a live cluster's etcd, and etcd is where Kubernetes keeps
its own state — so a bad write here is not an application-level mistake, it can
corrupt cluster state with no undo and nothing resembling a rollback. There is no
confirmation step beyond the editor: on save, the new value is put immediately.
Require explicit approval, and confirm which cluster is targeted first — the cluster
is the **first** argument, so a staging and a production invocation differ by one word.

**`edit` cannot be run unattended.** It launches `$EDITOR` (or `vim`) as a foreground
interactive process and waits for it, so an agent invoking it will hang. Only
`vim`, `nvim`, `nano`, `code`, `micro` and `emacs` are honoured — any other `$EDITOR`
is silently replaced by `vim`. Use `get` to read a value; there is no non-interactive
write path in this tree.

**`edit` corrupts values containing shell metacharacters, and can execute commands
inside the etcd pod.** The value is interpolated into an `echo "<value>" | etcdctl put`
string that is then evaluated by a shell **inside the pod**, with only newlines
escaped. So in the value being saved: `$FOO` is substituted with the pod's environment,
`` `cmd` `` is executed in the pod and its output substituted, `"` is stripped, and a
single quote `'` breaks the command outright with a syntax error. Verified by running
the constructed fragment through `sh`. Treat `edit` as unsafe for any value that is not
plain text — check the result with `get` afterwards rather than assuming the write was
faithful.

`edit` also writes the value to a scratch file at `<configPath>/<key>`, resolved against
the project root. That file contains real etcd contents and is left behind after
editing, so confirm the path is git-ignored. Because the key becomes part of the path, a
slash-style key like `/registry/config` creates nested directories under `configPath`,
while the flat filename-style keys the provider README uses (`cluster-prod.yaml`) do
not — both work, which is why the sample looks unlike a conventional etcd key.

`get` is read-only and safe for orientation, but note it prints the value to stdout —
if the key holds a credential, that reaches the transcript.

The root has no `Execute`, so this is not a passthrough: only `get` and `edit` exist,
and the rest of `etcdctl` — including anything that lists or deletes keys — is
unreachable from here.

#### Behaviour

Both verbs work by `kubectl exec`-ing into a **named pod** and running `etcdctl`
there; nothing connects to etcd directly. So the pod name in config must be an exact,
current pod name — not a deployment or selector — and a rescheduled pod with a new name
silently breaks both verbs until the config is updated. The `-it` flag means a TTY is
requested for every call, including `get`.

`<cluster>` completion is the intersection of two sources: clusters listed in this
provider's config **and** clusters `kubectl` knows a kubeconfig for. A cluster missing
from either side is not offered, though `get`/`edit` validate only against the config
and reject an unknown name with `invalid cluster`. `--profile` selects which kubectl
profile directory the kubeconfig is read from, and its completions come from the
directories under `kubectl`'s own `configPath`.

`<path>` completes from the cluster's configured `paths` list, but that list is
advisory — any key can be typed, and nothing checks that a typed key exists. `get` on a
missing key returns empty rather than failing, and `edit` on a missing key will happily
**create** it.

`edit` compares the file before and after editing and skips the write entirely when
nothing changed, reporting `no changes` — so an aborted edit is safe. Note it normalises
`\r\r\n` to `\n` when reading the value out of etcd, and the comparison is made against
that normalised text, so leaving the editor untouched really does write nothing; the
normalisation only affects what you are shown, not what is stored.

#### Configuration

Read under the `etcd` key by default. The override is `etcd.CommandWithConfigKey`, and
despite the name it is an `Option` on `etcd.New`, **not** on `NewCommand`. `NewCommand`
does declare a variadic `opts ...Option` parameter, but it never applies them and the
type targets `*ETCD` anyway — so anything passed there is silently ignored. There is
also no rename option: unlike every sibling, the command name is hardcoded as `etcd`.
Confirm against the project's own config file. Field shapes:
[`etcd-io/etcd/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/etcd-io/etcd/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`clusters` is the whole addressable surface, and each entry's `podName` and `namespace`
are what the exec targets — read them to know which pod a command will run in. Where
the kubeconfig comes from is governed by the `kubectl` provider's own `configPath`.

#### Examples

```bash
# Read a key. Read-only; prints the value. Keys come from the cluster's
# configured `paths` — the provider README uses filename-style keys.
posh execute etcd prod get cluster-prod.yaml

# Same, from a named kubectl profile.
posh execute etcd prod get cluster-prod.yaml --profile ci

# Interactive write — opens $EDITOR and puts on save. Needs a human
# and explicit approval; see Hazards.
posh execute etcd prod edit cluster-prod.yaml
```

#### References

- [`etcdctl get`/`put`](https://etcd.io/docs/latest/dev-guide/interacting_v3/)
- [etcd disaster recovery](https://etcd.io/docs/latest/op-guide/recovery/) — what a bad write may require
- [Provider README](https://github.com/foomo/posh-providers/blob/main/etcd-io/etcd/README.md) — config sample only; it documents no plugin wiring
