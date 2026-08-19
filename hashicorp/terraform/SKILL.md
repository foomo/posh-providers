#### Hazards

This tree changes live cloud infrastructure. The workspace is the **first**
argument, so a staging invocation and a production one differ by one word before
the verb — always confirm which workspace you are in before running anything that
writes, and prefer naming it explicitly over relying on shell history.

`destroy` tears down everything Terraform manages in the workspace. `apply`
creates, updates and **deletes** whatever the plan says, including replacing
resources in place. Both prompt for interactive confirmation by default and both
declare `--auto-approve`, which removes that last check — so
`destroy --auto-approve` is an unattended, irreversible teardown of a whole
environment. Never pass `--auto-approve` on behalf of a human, and require
explicit approval for `apply` and `destroy` even without it.

Omitting `[target]` **widens** these commands rather than narrowing them: with no
target, `apply` and `destroy` act on every resource in the workspace. The usage
block shows an optional repeated argument, which reads like less work, not more.

Supplying targets is not the safe middle ground it looks like either. They are
rewritten into `-target=` flags, and HashiCorp documents that option as being for
"exceptional circumstances only" — recovering from mistakes or working around
Terraform limitations — because a targeted run applies part of the configuration and
leaves the rest unreconciled, causing **undetected configuration drift**. So use
targets to reduce the blast radius of a deliberate repair, not as a routine way to
make `apply` less risky, and do not treat a successful targeted apply as evidence the
workspace matches its configuration.

`state rm` is the sharpest edge here and does not look like it. It removes
resources from Terraform's state **without touching the real infrastructure**, so
the resources keep existing and running while Terraform forgets it owns them — the
next `apply` will try to create duplicates, and the next `destroy` will leave the
originals behind. `state mv` can corrupt state the same way if the addresses are
wrong. Neither has an undo, and there is no confirmation prompt on either.

`unlock` force-releases the backend state lock. It is safe only when no run is
genuinely in progress; releasing a lock held by a live `apply` elsewhere invites two
concurrent writers and a corrupted state file. It declares `--force` to skip the
confirmation. Treat a stuck lock as a question for a human, not as an obstacle to
clear.

`import` writes to state, and `init` may reconfigure the backend. Safe for
orientation: `plan`, `validate`, `output`, `state list` and `state show` — these
read without writing, and `plan` is the correct way to find out what `apply` would
do.

Selecting `--service-principal` resolves a client secret out of the project config
into the child process's environment (`ARM_CLIENT_SECRET`), and it **overrides the
subscription too** — so the same workspace can be pointed at a different Azure
subscription by that flag alone. Check which credential is in play before trusting
that a workspace name implies a target account.

#### Behaviour

Workspaces are **discovered from the filesystem, not from config**. Any directory
under `path` that contains `.tf` files and is not prefixed with `.` or `_` is a
valid first argument, whether or not it has a `subscriptions` entry. A workspace
without one still runs — it just gets no `ARM_SUBSCRIPTION_ID` and, on `init`, no
`-backend-config` flags, so `init` silently falls back to whatever the `.tf` files
declare. The discovered list is cached per shell session, so a newly added workspace
directory needs a `cache clear` before it completes.

`[target]` completion is produced by **regex-scanning the workspace's `.tf` files**
for `module "x"` and `resource "type" "name"` declarations, not by asking Terraform.
So it lists addresses that are syntactically present rather than ones actually in
state, it ignores resources created by a module's internals, and it cannot see
`for_each`/`count` instance keys. A completed target is not proof the resource
exists.

Auth is assembled per invocation and depends on whether `--service-principal` was
given. With it, all four `ARM_*` variables come from that entry. Without it, only
`ARM_TENANT_ID` (if configured) and `ARM_SUBSCRIPTION_ID` (if the workspace has a
subscription entry) are set, and everything else falls through to the Azure
provider's own chain — `az login` locally, managed identity or OIDC in CI. So the
identity used can differ between two machines running the identical command.

`--service-principal` is declared on the *internal* flag set and is genuinely
consumed by posh: this provider forwards `fs.Default().Visited().Args()` rather than
the raw token list, so the flag does not leak to the `terraform` binary and only
flags the user actually typed are forwarded. Unlisted upstream flags go after `--`
and do reach the binary, on every verb.

`init` appends `-backend-config=` flags for resource group, storage account,
container and key, because Terraform's azurerm backend cannot read those from
environment variables. The key defaults to `<workspace>.tfstate` when unset. If any
of the three names is missing from the config, **no** backend flags are added at
all rather than a partial set.

`state` subcommands are dispatched by name into `terraform state <sub>`, and
`unlock` maps to upstream's `force-unlock` — the tree's name is not the binary's.
`Subscription.proxy` is read by nothing and has no effect.

#### Configuration

Read under the `terraform` key by default; `WithConfigKey` renames it and
`CommandWithName` renames the command — both are `CommandOption`s on `NewCommand`
here, unlike several sibling providers that split them across two constructors.
`CommandWithMiddlewares` wraps every exec. Confirm against the project's own config
file. Field shapes:
[`hashicorp/terraform/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/hashicorp/terraform/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`servicePrincipals` holds plaintext client secrets, so that file is credential
material — check it is git-ignored before writing to it, and read `subscriptions` to
learn which Azure subscription each workspace is expected to target.

#### Examples

```bash
# Orientation, in increasing order of commitment. All read-only.
posh execute terraform staging validate
posh execute terraform staging state list
posh execute terraform staging plan

# Narrow a plan to one module. Fine for inspection; see Hazards before
# carrying the same target into an apply.
posh execute terraform staging plan module.network

# Apply, still interactive — terraform asks before writing.
posh execute terraform staging apply

# Read an output value.
posh execute terraform staging output --json

# Use an explicit service principal instead of the ambient Azure login.
posh execute terraform staging plan --service-principal ci
```

#### References

- [Terraform CLI commands](https://developer.hashicorp.com/terraform/cli/commands)
- [`terraform state rm`](https://developer.hashicorp.com/terraform/cli/commands/state/rm)
- [`terraform force-unlock`](https://developer.hashicorp.com/terraform/cli/commands/force-unlock)
- [azurerm backend](https://developer.hashicorp.com/terraform/language/backend/azurerm)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/hashicorp/terraform/README.md)
