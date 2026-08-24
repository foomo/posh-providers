#### Hazards

`up`, `destroy`, `refresh`, `import` and `state` mutate live Azure infrastructure and
**must not be run unattended**. The target comes from `<env> <verb> <project>
<stack>`, so a local invocation and a production one differ by one word. The backend's
`subscription` becomes `ARM_SUBSCRIPTION_ID`, deciding where resources are created.

`destroy` deletes every resource in the stack; `--target` narrows it and
`--exclude-protected` spares protected ones, but neither is default. `--remove` also
deletes the stack and its config file, so the state is gone after.

`state delete`, `state rename` and `state unprotect` edit state **without touching
the cloud** — the dangerous direction, since the record then disagrees with reality and
a later `up` can recreate or orphan resources. Read-only: `stack`, `preview`, `cancel`
— but `stack output` prints resolved outputs, routinely including secrets.

**Every stack verb resolves a 1Password secret and fetches a storage account key
first**, so all need an unlocked session plus a signed-in `az`. A `Pulumi.<stack>.op`
file beside the stack yaml is piped through `op inject` and written into live stack
config with `config set-all --secret` *before* the verb runs — so `preview`, nominally
read-only, still mutates stack configuration. Those values reach an `sh -c` line
**unquoted**: a secret containing `$(...)`, backticks or a space is corrupted or
executed.

The storage account key is fetched fresh on **every** stack operation and exported as
`AZURE_STORAGE_KEY` — an account-wide credential granting full access to the whole
storage account, not just the state container.

`backend create` creates a real resource group, storage account and container, and
splits `--group-args`/`--storage-args` on spaces into the `az` lines — arbitrary flags
reaching a resource-creating call.

`PULUMI_HOME` is repointed at `configPath` process-wide, relocating credentials and
plugins into the project rather than the home directory.

#### Behaviour

`<env>` is **not validated against config** — it completes from directory names under
`path`, so an env in `backends:` but missing from disk fails with `invalid command`,
while a directory with no entry completes then fails with `backend not found`.
`<stack>` completes from `Pulumi.<name>.yaml` files, so one existing only in the
backend is never suggested.

#### Configuration

Config key `pulumi`, shared with the gcloud sibling through a schema `oneOf`; both
default to the same command name, so a project wires one. Field shapes:
[`pulumi/pulumi/azure/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/pulumi/pulumi/azure/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} prod stack myproject live output
posh execute {{cmd}} prod up myproject live
```

#### References

- [pulumi CLI](https://www.pulumi.com/docs/iac/cli/commands/)
- [Azure Blob backend](https://www.pulumi.com/docs/iac/concepts/state-and-backends/#azure-blob-storage)
