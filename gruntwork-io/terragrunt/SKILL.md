#### Hazards

This tree changes live cloud infrastructure, and the target is chosen by the **first
two** arguments — staging and production differ by one word, two positions before the
verb. Confirm both before running anything that writes.

`destroy` tears down each named stack; `apply` creates, updates and **deletes**
whatever the plan says. Neither gets `-auto-approve` here, so the upstream
confirmation applies — but it prompts **per stack**. Require approval for both.

**`<stacks>` is required and repeated, and each name is a separate run.** They run
sequentially and abort at the first failure, so `apply net/vpc db/main` can leave the
first applied, the second failed and the rest untouched under one non-zero exit.

**Giving `<env>` and `<site>` with no verb crashes.** The `site` node has its own
`Execute` but declares no arguments, so nothing checks the count before the handler
slices from index 3 — a `slice bounds out of range` panic, not a usage error.

`secrets` renders every `secrets.tpl.yaml` under the site to a plaintext
`secrets.yaml` beside it via 1Password. Those files land in the checkout and
**nothing removes them**, so confirm they are git-ignored. Needs a live 1Password
session, and gates any stack reading them.

Read-only: `validate`, `plan`, `output`. There is no `state`, `import` or
`force-unlock` verb — further verbs are commented out in the source, so the rest of
the upstream CLI is unreachable, not merely unlisted.

#### Behaviour

The command surface is **discovered from the filesystem**, not configured: envs and
sites are the non-dotted, non-underscored directories under `<path>/envs`, stacks come
from walking the site for `terragrunt.hcl`. A `terragrunt.hcl` sitting directly in the
site directory never completes, though one can still be typed. Cached per session.

**`output --raw` is declared then discarded — the one flag this tree offers cannot
reach the binary**, because only `AdditionalFlags()` is forwarded and it strips every
declared flag. An *undeclared* one passes through. Args beyond the stack list and
anything after `--` are dropped.

The `secrets` templates use Go `text/template` with `<% %>` delimiters, not
`op inject` syntax.

#### Configuration

Config key `terragrunt`. Field shapes:
[`gruntwork-io/terragrunt/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/gruntwork-io/terragrunt/config.schema.json).

`path` alone decides what this command reaches. The README misspells `cachePath` as
`cachPath`, so a project copying it downloads modules into the project root.

#### Examples

```bash
posh execute {{cmd}} prod eu plan net/vpc
posh execute {{cmd}} prod eu apply net/vpc net/subnets
```

#### References

- [Terragrunt CLI reference](https://terragrunt.gruntwork.io/docs/reference/cli-options/)
- [`onepassword` provider](https://github.com/foomo/posh-providers/tree/main/onepassword) — `secrets.tpl.yaml` functions
