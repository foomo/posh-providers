#### Hazards

`up`, `destroy`, `refresh`, `import` and `state` mutate live Google Cloud
infrastructure and **must not be run unattended**. The target comes from `<env>
<verb> <project> <stack>`, so a local invocation and a production one differ by one
word and nothing in the tree marks which is which.

`destroy` deletes every resource in the stack; `--target` narrows it and
`--exclude-protected` spares protected ones, but neither is the default. `--remove`
also deletes the stack and its config file, so the state is gone after.

`state delete`, `state rename` and `state unprotect` edit state **without touching
the cloud** — the dangerous direction, since the record then disagrees with reality
and a later `up` can recreate or orphan resources. Read-only: `stack`, `preview` and
`cancel` — but `stack output` prints resolved outputs, routinely including secrets.

**Every stack verb resolves a 1Password secret first**, so all need an unlocked
session. A `Pulumi.<stack>.op` file beside the stack yaml is piped through
`op inject` and written into live stack config with `config set-all --secret`
*before* the verb runs — so `preview`, nominally read-only, still mutates stack
configuration. Those values reach an `sh -c` line **unquoted**: a secret containing
`$(...)`, backticks or a space is corrupted or executed.

`backend create` creates a real GCS bucket. `backend login` runs `gcloud auth
application-default login`, a **browser flow that cannot complete unattended**, then
writes credentials to `devops/config/gcloud/application_default_credentials.json`,
which every stack verb expects to exist — so a fresh checkout needs a human to run it
first, and that file must never be committed.

`PULUMI_HOME` is repointed at `configPath` process-wide, relocating credentials and
plugins into the project rather than the user's home directory.

#### Behaviour

`<env>` is **not validated against config** — it completes from directory names under
`path`, so an env in `backends:` but missing from disk fails with `invalid command`,
while a directory with no `backends:` entry completes then fails with `backend not
found`. `<stack>` completes from `Pulumi.<name>.yaml` files, so a stack existing only
in the backend is never suggested, though typing it works.

#### Configuration

Config key `pulumi`, shared with the azure sibling through a schema `oneOf`; both
default to the same command name, so a project wires one. Field shapes:
[`pulumi/pulumi/gcloud/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/pulumi/pulumi/gcloud/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} prod stack myproject live output
posh execute {{cmd}} prod up myproject live
```

#### References

- [pulumi CLI](https://www.pulumi.com/docs/iac/cli/commands/)
- [Google Cloud Storage backend](https://www.pulumi.com/docs/iac/concepts/state-and-backends/#google-cloud-storage)
