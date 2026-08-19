#### Hazards

This tree changes live cloud infrastructure through Terraform, one stack at a time.

`destroy` tears down everything in the named stacks. `deploy` creates, updates and
**deletes** whatever the synthesized plan says. Both prompt for interactive approval
by default and both declare `--auto-approve`, which removes that check — so
`destroy --auto-approve` is an unattended, irreversible teardown. Never pass
`--auto-approve` on behalf of a human, and require explicit approval for `deploy`
and `destroy` even without it.

`[stacks]` is optional and **omitting it widens the command to every stack in the
app**, not to none. The usage block shows an optional repeated argument, which reads
like less work. Always name the stacks for `deploy` and `destroy`; `cdktf destroy`
with no argument is the whole application.

`unlock` force-releases a stack's Terraform state lock by running `terraform
force-unlock` directly in the synthesized output directory. It is safe only when no
run is genuinely in progress — releasing a lock held by a live `deploy` invites two
concurrent writers and a corrupted state file. Treat a stuck lock as a question for
a human. Note it needs `outPath` configured; without it the command runs in the
wrong directory and fails.

`--migrate-state` on `deploy`/`destroy`/`diff` tells Terraform to copy existing state
into a newly configured backend. It is meant for a deliberate backend switch, not as
a way to clear an unexpected error, and it normally prompts before migrating each
workspace's state — so do not pass it speculatively, and expect it to block waiting
for that confirmation.

Safe for orientation: `list` and `diff`. `diff` is the correct way to find out what
`deploy` would do. The root has no `Execute`, so this is **not** a passthrough:
unmatched subcommands are rejected and the rest of the cdktf CLI — `synth`, `get`,
`output`, `watch`, `convert`, `provider` — is unreachable from here.

#### Behaviour

**`--skip-synth` behaves differently on all four verbs that declare it, and does
nothing on the one where it looks most carefully handled.** The provider translates
the flag into a `SKIP_SYNTH=true` environment variable, but cdktf reads no such
variable — its yargs setup binds only the `CDKTF_` prefix, so `SKIP_SYNTH` is
invisible to it. What actually decides the outcome is whether the verb also forwards
the raw flag:

- `diff` and `destroy` forward `r.Flags()`, which includes internal flags, so
  `--skip-synth` reaches cdktf and **works** — it is a genuine cdktf flag on both.
- `deploy` forwards only `fs.Visited().Args()` (the default set), so the flag is
  stripped and nothing but the dead env var remains: on `deploy`, `--skip-synth`
  has **no effect** and the app is re-synthesized anyway.
- `list` forwards it too, but cdktf's `list` has no such flag. cdktf is not run in
  yargs strict mode for these commands, so it is **silently ignored** rather than
  rejected.

So the flag works on the two verbs that leak it by accident and is inert on the one
that handles it deliberately. Do not rely on it for `deploy`; pass it after `--` if
you need it there.

Stack names are discovered by walking `path` for **`*.stack.yaml` files** and
stripping that suffix from the basename, not by asking cdktf. So the completion list
reflects a file-naming convention rather than the stacks the app actually defines —
a stack without a matching `*.stack.yaml` cannot be named or completed, and a stray
file produces a stack name that does not exist. `CommandWithStackNameProvider`
replaces the derivation, so in a project that passes it the mapping may differ
entirely. The list is cached per shell session, so a newly added stack file needs a
`cache clear`.

Log verbosity is wired to posh's own level at construction time: `CDKTF_LOG_LEVEL`
is exported as `debug` at trace and `info` at debug. It is set once for the process,
so it affects every cdktf invocation in the session, and `list`/`diff`/`destroy`
forward unlisted flags verbatim.

`unlock` does not call cdktf at all — it invokes `terraform force-unlock` in
`<outPath>/stacks/<stack>`, so it depends on a prior synth having produced that
directory and on `terraform` being on PATH. Everything else requires the `cdktf`
binary, which `Validate` checks for, suggesting `npm install --global cdktf-cli`.

#### Configuration

Read under the `cdktf` key by default; `WithConfigKey` renames it and
`CommandWithName` renames the command — both are `CommandOption`s on `NewCommand`.
Confirm against the project's own config file. Field shapes:
[`hashicorp/cdktf/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/hashicorp/cdktf/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

Both fields matter operationally: `path` is the working directory for every verb and
the root stacks are discovered under, and `outPath` is required by `unlock` alone.
The provider README documents only `path` and its plugin snippet is a copy of
`helm`'s, so do not take either from it.

#### Examples

```bash
# Orientation, read-only.
posh execute cdktf list
posh execute cdktf diff my-stack

# Deploy one named stack. Still interactive; cdktf asks before writing.
posh execute cdktf deploy my-stack

# Reuse an existing synth on diff/destroy, where the flag actually applies.
posh execute cdktf diff my-stack --skip-synth

# Release a stuck lock. Needs outPath configured; confirm no run is active.
posh execute cdktf unlock my-stack 9f8e7d6c-0000-0000-0000-000000000000
```

#### References

- [CDKTF CLI reference](https://developer.hashicorp.com/terraform/cdktf/cli-reference/commands)
- [`terraform force-unlock`](https://developer.hashicorp.com/terraform/cli/commands/force-unlock)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/hashicorp/cdktf/README.md) — its plugin snippet calls `helm.NewCommand` and does not apply to this provider
