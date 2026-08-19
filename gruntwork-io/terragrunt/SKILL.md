#### Hazards

This tree changes live cloud infrastructure, and the target is chosen by the **first
two** arguments — so a staging run and a production one differ by one word, two
positions before the verb. Confirm both `<env>` and `<site>` before running anything
that writes.

`destroy` tears down the infrastructure in each named stack. `apply` creates, updates
and **deletes** whatever the plan says. Neither is given `-auto-approve` by the
provider, so terragrunt's own interactive confirmation still applies — but note it
prompts **per stack**, and the loop below means a partial run is possible. Require
explicit approval for both.

**`<stacks>` is a required, repeated argument and each name is a separate terragrunt
run.** The command loops over them and executes sequentially, aborting at the first
failure — so `apply net/vpc net/subnets db/main` can leave the first stack applied,
the second failed and the third untouched, with a single non-zero exit covering all
three. Read the logged stack list to see how far it got; the exit status alone does not
say.

**`terragrunt <env> <site>` with no verb crashes.** The `site` node carries its own
`Execute` but declares no arguments, so nothing validates the arg count before the
handler slices from index 3 — the result is a `slice bounds out of range [3:2]` panic
rather than a usage message. Always give a verb. Verified by running that exact
invocation through the command.

`secrets` renders every `secrets.tpl.yaml` under the site to a plaintext
`secrets.yaml` beside it, resolving values through 1Password. Those files are written
into the checkout with a `DO NOT EDIT` header and **nothing removes them**, so confirm
they are git-ignored and require a working 1Password session — without one the render
fails. It is the only verb here that touches 1Password, and it is a prerequisite for
any stack whose configuration reads those files.

Note the templates are **not** `op inject` templates: this provider renders them with
Go's `text/template` using `<% %>` delimiters and an `op` function, so the syntax is
posh's own, not 1Password's, and `missingkey=error` means an unresolved reference fails
the render rather than emitting a blank.

Safe for orientation: `validate`, `plan` and `output`. `plan` is the correct way to
find out what `apply` would do. There is no `state`, `import` or `force-unlock` verb —
a large block of additional terragrunt commands is present but commented out in the
source, so the rest of the CLI is unreachable from here rather than merely unlisted.

#### Behaviour

The entire command surface is **discovered from the filesystem**, not configured. Envs
are the non-dotted, non-underscored directories under `<path>/envs`; sites the same
one level deeper; stacks come from walking the site for `terragrunt.hcl` files. So a
directory prefixed `.` or `_` is deliberately invisible, and adding an env, site or
stack changes what completes with no config edit. All three lists are cached per shell
session, so new directories need a `cache clear`.

Stack discovery keeps only paths containing a `/` after the site prefix is trimmed —
so a `terragrunt.hcl` sitting **directly** in the site directory is dropped from
completion, and only nested stacks like `net/vpc` are offered. A top-level stack can
still be run by typing its name.

**`output --raw` is declared and then discarded — the one flag this tree offers is the
one flag that cannot reach the binary.** The provider forwards `AdditionalFlags()`
only, and that accessor returns the typed flags *minus every flag declared on a flag
set*, stripping two tokens for a value flag. Since `output` declares `--raw`, typing it
removes it: terragrunt is invoked with no flag at all and prints the normal output
rather than the raw string. An **undeclared** flag does pass through, so
`output net/vpc --json` reaches terragrunt while `--raw vpc_id` does not. Verified by
running both through the parser with this leaf's flag set registered.

Nothing else is forwarded either: `Args()` beyond the stacks list and
`AdditionalArgs()` are both dropped, so arguments after `--` never reach terragrunt.
Practically, undeclared flags typed directly on the verb are the only extra input this
command accepts.

Each stack runs with the stack directory as its working directory, and the verb is
passed to terragrunt verbatim as its first argument. `TERRAGRUNT_DOWNLOAD` is exported
once at shell startup from `cachePath`, so it applies to every terragrunt invocation in
the session, including ones typed outside this command.

#### Configuration

Read under the `terragrunt` key by default; `WithConfigKey` renames it and
`CommandWithName` renames the command, both `CommandOption`s on `NewCommand`. Confirm
against the project's own config file. Field shapes:
[`gruntwork-io/terragrunt/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/gruntwork-io/terragrunt/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`path` is the only knob that decides what this command can reach — read it, then look
at `<path>/envs` to see the real environment and site names rather than trusting
completion. Note the provider README misspells `cachePath` as **`cachPath`**, so a
project configured from it silently has no cache path and terragrunt downloads modules
into the project root; check the spelling in the project's own file.

#### Examples

```bash
# Orientation, read-only. Always give a verb — env+site alone panics.
posh execute terragrunt prod eu validate net/vpc
posh execute terragrunt prod eu plan net/vpc

# Materialise secrets first if the stacks read them. Needs 1Password.
posh execute terragrunt prod eu secrets

# Apply one stack. Terragrunt still prompts for confirmation.
posh execute terragrunt prod eu apply net/vpc

# Several stacks run sequentially and stop at the first failure.
posh execute terragrunt prod eu apply net/vpc net/subnets

# Read outputs. Note --raw is declared but discarded before terragrunt sees it,
# so this prints the full output block; see Behaviour.
posh execute terragrunt prod eu output net/vpc
```

#### References

- [Terragrunt CLI reference](https://terragrunt.gruntwork.io/docs/reference/cli-options/)
- [Terragrunt configuration](https://terragrunt.gruntwork.io/docs/reference/config-blocks-and-attributes/)
- [`onepassword` provider](https://github.com/foomo/posh-providers/tree/main/onepassword) — the template functions available in `secrets.tpl.yaml`
- [Provider README](https://github.com/foomo/posh-providers/blob/main/gruntwork-io/terragrunt/README.md) — its config sample misspells `cachePath`
