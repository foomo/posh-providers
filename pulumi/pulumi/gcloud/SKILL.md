#### Hazards

`up`, `destroy`, `refresh`, `import` and `state` mutate live Google Cloud
infrastructure through pulumi and **must not be run unattended**. The target is
chosen by the first three arguments — `<env> <verb> <project> <stack>` — so a
local invocation and a production one differ by one word, and nothing in the tree
marks which environment is which. Confirm the environment and stack against the
project's own config before running any of them.

`destroy` deletes every resource in the stack, not a subset; `--target` narrows
it and `--exclude-protected` spares protected resources, but neither is the
default. `--remove` additionally deletes the stack and its config file, so the
state itself is gone afterwards.

`state delete`, `state rename` and `state unprotect` edit the state file without
touching the cloud, which is the dangerous direction: they make pulumi's record
disagree with reality, so a later `up` can recreate or orphan resources.
`state upgrade` migrates the backend's on-disk format. Only `stack` (`init`,
`output`, `history`), `preview` and `cancel` are safe for orientation — and
`stack output` prints resolved outputs, which routinely include secrets.

**Every stack verb resolves a 1Password secret before doing anything**, so all of
them need an unlocked session (`onepassword` provider) and fail otherwise. Where
a `Pulumi.<stack>.op` file sits beside the stack yaml, it is piped through
`op inject` and its lines are written into live stack config with
`pulumi config set-all --secret` *before* the requested verb runs — so `preview`,
nominally read-only, still mutates the stack's configuration. Those values are
interpolated into an `sh -c` line unquoted: a secret containing `$(...)`,
backticks or a space is corrupted or executed, and an unbalanced quote aborts the
command. Keep injected values free of shell metacharacters.

`backend create` creates a real GCS bucket and `backend login` runs
`gcloud auth application-default login`, a **browser flow** that cannot complete
unattended. It then writes application default credentials to
`devops/config/gcloud/application_default_credentials.json` in the project, which
every stack verb expects to exist — so a fresh checkout needs `backend login`
run by a human first, and that file must never be committed.

`PULUMI_HOME` is repointed at `configPath` process-wide when the command is
constructed, so pulumi's credentials and plugins live in the project rather than
the user's home directory.

#### Behaviour

`<env>` is not validated against config — it is completed from the directory
names under `path`. An environment present in `backends:` but absent from disk
never matches, so the invocation falls through to the root, which has no
`Execute`, and fails with `invalid command` rather than naming the real problem.
The reverse also holds: a directory with no matching `backends:` entry completes
and then fails with `backend not found`.

`<project>` completes from directories below `<env>`, and `<stack>` from
`Pulumi.<name>.yaml` files inside the project — so a stack that exists in the
backend but has no local yaml is never suggested, though typing it works. All
arguments are required; omitting one is rejected as `missing argument` before
anything runs.

The verb is taken from the node name and forwarded to `pulumi` verbatim, and the
command runs with the project directory as its working directory. A repeated
flag has its second and later occurrences forwarded twice (`--target a --target
b` sends `b` twice), because only the first occurrence of each flag name is
removed before the remainder is appended. Harmless for `--target`, which pulumi
treats as a set, but the forwarded command line is not what was typed. A `--`
separator is itself forwarded as a positional argument.

`backend create` declares `--debug`, `--tags` and `--vebose` as *value-taking*
strings, and none of the three is read by its `Execute` — it passes only
`--location` and `--project`. The flag name `vebose` is misspelled, and the
azure sibling declares the same three as booleans. The `StorageAccount` type in
`storageaccount.go` is referenced by nothing.

#### Configuration

Default config key `pulumi`, overridable with `CommandWithConfigKey`; confirm
against the project's own config file. Field shapes:
[`pulumi/pulumi/gcloud/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/pulumi/pulumi/gcloud/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`. Note the `pulumi` key is a
`oneOf` over this provider and its azure sibling, which share the key and the
command name — a project wires one of the two.

#### Examples

```bash
# Safe: inspect a stack without touching it
posh execute pulumi prod stack myproject live output
posh execute pulumi prod preview myproject live

# Mutating: confirm the environment first
posh execute pulumi prod up myproject live
posh execute pulumi prod up myproject live --target 'urn:pulumi:live::myproject::gcp:storage/bucket:Bucket::assets'
```

#### References

- [pulumi CLI](https://www.pulumi.com/docs/iac/cli/commands/)
- [pulumi Google Cloud Storage backend](https://www.pulumi.com/docs/iac/concepts/state-and-backends/#google-cloud-storage)
- [provider README](https://github.com/foomo/posh-providers/blob/main/pulumi/pulumi/gcloud/README.md) — its title says *azure* and its plugin snippet passes the constructor arguments in the wrong order; the real signature is `NewCommand(l, gcloud, op, cache)`.
