#### Hazards

`edit` writes directly into a live cluster's key-value store, where Kubernetes keeps its
own state — so a bad write can corrupt cluster state with no undo. There is no
confirmation beyond the editor: on save the value is written immediately. Require explicit
approval, and confirm the target — the cluster is the **first** argument, so staging and
production differ by one word.

**`edit` cannot be run unattended.** It launches `$EDITOR` (or `vim`) in the foreground
and waits, so an agent invoking it hangs. Only `vim`, `nvim`, `nano`, `code`, `micro` and
`emacs` are honoured — any other `$EDITOR` is silently replaced by `vim`. There is no
non-interactive write path here.

**`edit` corrupts values containing shell metacharacters, and can execute commands inside
the pod.** The value is interpolated into an `echo` pipeline evaluated by a shell **inside
the pod**, with only newlines escaped: `$FOO` is substituted from the pod's environment, a
backtick command is executed there, `"` is stripped, and a single quote breaks the command
outright. Treat `edit` as unsafe for anything but plain text, and verify with `get` after.

`edit` also leaves the value in a scratch file under `configPath` inside the project,
holding real cluster data — confirm that path is git-ignored.

`get` is read-only, but it prints the value to stdout — if the key holds a credential,
that reaches the transcript.

#### Behaviour

Both verbs `kubectl exec` into a **named pod** and run the client there. The configured
pod name must be exact and current — not a deployment or selector — so a rescheduled pod
silently breaks both verbs until the config is updated.

The key argument completes from the cluster's configured `paths`, but that list is
advisory — any key can be typed and nothing checks it exists. `get` on a missing key
returns empty rather than failing, and `edit` on a missing key **creates** it.

#### Configuration

Read under the `etcd` key. Unlike every sibling the command name is hardcoded.
`clusters` is the whole addressable surface, and each entry's `podName` and `namespace`
are what the exec targets.

Field shapes: [`etcd-io/etcd/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/etcd-io/etcd/config.schema.json).

#### Examples

```bash
# Read a key. Read-only; prints the value.
posh execute {{cmd}} prod get cluster-prod.yaml --profile ci

# Interactive write - needs a human and explicit approval.
posh execute {{cmd}} prod edit cluster-prod.yaml
```

#### References

- [`etcdctl` interacting](https://etcd.io/docs/latest/dev-guide/interacting_v3/)
- [Disaster recovery](https://etcd.io/docs/latest/op-guide/recovery/) — what a bad write may require
