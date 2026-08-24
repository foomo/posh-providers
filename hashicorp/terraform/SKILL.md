#### Hazards

This tree changes live cloud infrastructure, and the workspace is the **first**
argument — staging and production differ by one word before the verb. Confirm it
before running anything that writes.

`apply` creates, updates and **deletes** whatever the plan says; `destroy` tears
down everything Terraform manages in the workspace. Both prompt interactively and
both declare `--auto-approve`, which removes that last check. Never pass it on
behalf of a human, and require explicit approval for both regardless.

Omitting `[target]` **widens** these commands rather than narrowing them: with no
target, `apply` and `destroy` act on every resource in the workspace. Targets are no
safe middle ground either — a targeted run leaves the rest unreconciled, causing
drift.

`state rm` is the sharpest edge and does not look like it: it drops resources from
state **without touching the real infrastructure**, so they keep running while
Terraform forgets it owns them — the next `apply` creates duplicates, the next
`destroy` leaves the originals. `state mv` corrupts state the same way on a wrong
address. Neither prompts, neither has an undo.

`unlock` force-releases the backend state lock; releasing one held by a live `apply`
elsewhere invites two concurrent writers and a corrupted state file. Treat a stuck
lock as a question for a human.

`import` writes state; `init` may reconfigure the backend. Read-only: `plan`,
`validate`, `output`, `state list`/`show`.

`--service-principal` resolves a plaintext client secret from config into the child
environment as `ARM_CLIENT_SECRET`, and **overrides the subscription too** — the
same workspace can be pointed at a different Azure subscription by that flag alone.

#### Behaviour

Workspaces are **discovered from the filesystem, not config**: any directory under
`path` holding `.tf` files and not prefixed `.` or `_`. One without a
`subscriptions` entry still runs, just with no `ARM_SUBSCRIPTION_ID` and, on `init`,
no `-backend-config` flags. Cached per session.

`[target]` completion **regex-scans the workspace's `.tf` files** rather than asking
Terraform, so it lists addresses that are syntactically present, not ones in state,
and cannot see `for_each`/`count` keys.

Without `--service-principal` the credential falls through to the Azure provider's
own chain, so the identity can differ between machines.

#### Configuration

Config key `terraform`. Field shapes:
[`hashicorp/terraform/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/hashicorp/terraform/config.schema.json).

`servicePrincipals` holds plaintext client secrets — treat that file as credential
material.

#### Examples

```bash
posh execute {{cmd}} staging plan
posh execute {{cmd}} staging apply
posh execute {{cmd}} staging plan --service-principal ci
```

#### References

- [Terraform CLI commands](https://developer.hashicorp.com/terraform/cli/commands)
- [azurerm backend](https://developer.hashicorp.com/terraform/language/backend/azurerm)
