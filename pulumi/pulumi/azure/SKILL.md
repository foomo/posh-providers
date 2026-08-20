#### Hazards

`up`, `destroy`, `refresh`, `import` and `state` mutate live Azure
infrastructure through pulumi and **must not be run unattended**. The target is
chosen by the first three arguments — `<env> <verb> <project> <stack>` — so a
local invocation and a production one differ by one word, and nothing in the tree
marks which environment is which. Confirm the environment and stack against the
project's own config before running any of them. The backend's `subscription` is
exported as `ARM_SUBSCRIPTION_ID`, so it also decides which subscription the
stack's resources are created in.

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

**Every stack verb resolves a 1Password secret and fetches a storage account key
before doing anything**, so all of them need an unlocked session (`onepassword`
provider) plus a signed-in `az` CLI, and fail otherwise. Where a
`Pulumi.<stack>.op` file sits beside the stack yaml, it is piped through
`op inject` and its lines are written into live stack config with
`pulumi config set-all --secret` *before* the requested verb runs — so `preview`,
nominally read-only, still mutates the stack's configuration. Those values are
interpolated into an `sh -c` line unquoted: a secret containing `$(...)`,
backticks or a space is corrupted or executed, and an unbalanced quote aborts the
command. Keep injected values free of shell metacharacters.

The storage account key is fetched with `az storage account keys list` on every
stack operation and passed to pulumi as `AZURE_STORAGE_KEY`. It is an
account-wide credential granting full access to the whole storage account, not
just the state container, and it is read fresh each time rather than cached.

`backend create` creates a real resource group, storage account and container,
and its `--group-args`/`--storage-args` values are split on spaces and appended
to the `az` command lines — arbitrary flags reaching a resource-creating call.
`backend login` needs the storage key too, so it needs `az` already signed in.

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

`backend create`'s `--tags`, `--debug` and `--vebose` are declared but never
read; only `--group-args`/`--storage-args` and the visited default flags reach
`az`. The flag name `vebose` is misspelled. The `StorageAccount` type in
`storageaccount.go` is referenced by nothing.

#### Configuration

Default config key `pulumi`, overridable with `CommandWithConfigKey`; confirm
against the project's own config file. Field shapes:
[`pulumi/pulumi/azure/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/pulumi/pulumi/azure/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`. Note the `pulumi` key is a
`oneOf` over this provider and its gcloud sibling, which share the key and the
command name — a project wires one of the two.

#### Examples

```bash
# Safe: inspect a stack without touching it
posh execute pulumi prod stack myproject live output
posh execute pulumi prod preview myproject live

# Mutating: confirm the environment first
posh execute pulumi prod up myproject live
posh execute pulumi prod up myproject live --target 'urn:pulumi:live::myproject::azure-native:storage:StorageAccount::assets'
```

#### References

- [pulumi CLI](https://www.pulumi.com/docs/iac/cli/commands/)
- [pulumi Azure Blob Storage backend](https://www.pulumi.com/docs/iac/concepts/state-and-backends/#azure-blob-storage)
- [provider README](https://github.com/foomo/posh-providers/blob/main/pulumi/pulumi/azure/README.md) — its plugin snippet passes the constructor arguments in the wrong order; the real signature is `NewCommand(l, az, op, cache)`.
