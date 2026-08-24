#### Hazards

Both verbs are agent-hostile; neither can be completed unattended.

`auth` opens a browser and no flag turns it headless, so an agent invoking it
hangs. The non-interactive path is not wired into this tree: run the upstream
`activate-service-account` verb outside posh (service account key, token, or
OIDC) and every command here works against that session. Ask a human rather than
retrying `auth`.

`kubeconfig` cannot run unattended either. Upstream prompts for confirmation, and
the `--assume-yes` that would skip it cannot be delivered: posh keeps the literal
`--`, so what follows arrives as a positional and the verb accepts exactly one.
`kubeconfig dev -- --assume-yes` fails with an argument-count error instead of
auto-confirming.

It also writes into the checkout, since the kubectl config path resolves against
the project root - confirm that path is git-ignored. Without `--login`, never
passed here, the file embeds no secret and calls the CLI for short-lived
credentials, so it only works where the CLI is installed and logged in.

Nothing here deletes or reconfigures cloud resources, and the root has no
`Execute`, so the rest of the CLI is unreachable.

#### Behaviour

What expires is the CLI login session, not the file: the login the kubeconfig
depends on defaults to 2 hours. So a `kubectl` call that worked earlier can start
failing on authentication with the file unchanged - read that as an expired login
needing `auth` again, not a file to regenerate.

The provider passes no `--overwrite`, so upstream merges into the target file.
Each alias normally gets its own `<alias>.yaml`, hiding this; but a
`CommandWithClusterNameFn` mapping two aliases onto one filename accumulates both
as separate contexts, and the current context - never set here - then decides
which cluster `kubectl` talks to.

`<project>` and `<cluster>` are local aliases, not STACKIT names, completing from
config map keys. A project's `id` is the UUID passed as `--project-id` and a
cluster's `name` is the real SKE cluster, so the alias you type is not what the
CLI receives. An unknown one is rejected up front.

#### Configuration

The `stackit` key's `projects` map is the whole addressable surface - a project or
cluster absent from it cannot be reached. Where the kubeconfig lands is governed
by the `kubectl` provider.

Field shapes: [`stackitcloud/stackit/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/stackitcloud/stackit/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} project my-project kubeconfig dev --profile ci
```

#### References

- https://docs.stackit.cloud/developer-tools/stackit-cli/
- https://github.com/stackitcloud/stackit-cli/blob/main/docs/stackit_auth_activate-service-account.md
