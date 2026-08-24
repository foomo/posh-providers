#### Hazards

This tree changes live cloud infrastructure through Terraform.

`destroy` tears down everything in the named stacks; `deploy` creates, updates and
**deletes** whatever the synthesized plan says. Both prompt for approval and both
declare `--auto-approve`, which removes that check. Never pass it on behalf of a
human, and require explicit approval for both regardless.

`[stacks]` is optional and **omitting it widens the command to every stack in the
app**, not to none — the usage block reads like less work. Always name the stacks for
`deploy` and `destroy`.

`unlock` force-releases a stack's Terraform state lock. Releasing one held by a live
`deploy` invites two concurrent writers and a corrupted state file — treat a stuck
lock as a question for a human. Needs `outPath`, or it runs in the wrong directory.

`--migrate-state` copies existing state into a newly configured backend — for a
deliberate backend switch, not to clear an error. It prompts, so expect it to block.

Read-only: `list` and `diff`. `diff` is the correct way to find out what `deploy`
would do. The root carries no `Execute`, so this is **not** a passthrough — the rest
of the upstream CLI is unreachable.

#### Behaviour

**`--skip-synth` reaches the binary on three verbs and is silently dropped on
`deploy`.** The provider turns it into a `SKIP_SYNTH=true` env var on all four, but
only `list`, `diff` and `destroy` also forward the raw flag; `deploy` forwards just
the visited default flags, so the app is re-synthesized anyway. Pass it after `--`
if you need it there.

Stack names come from walking `path` for **`*.stack.yaml` files**, not from querying
the app — so completion reflects a file-naming convention rather than the stacks
actually defined: a stack without a matching file cannot be completed, and a stray
file yields one that does not exist. `CommandWithStackNameProvider` replaces that
derivation. Cached per session.

`CDKTF_LOG_LEVEL` is exported once at construction from posh's own level.

`unlock` calls no `cdktf` at all — it runs `terraform` in `<outPath>/stacks/<stack>`,
so it needs a prior synth and `terraform` on PATH.

#### Configuration

Config key `cdktf`. Field shapes:
[`hashicorp/cdktf/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/hashicorp/cdktf/config.schema.json).

`path` is the working directory for every verb; `outPath` is used by `unlock` only.

#### Examples

```bash
posh execute {{cmd}} diff my-stack
posh execute {{cmd}} deploy my-stack
posh execute {{cmd}} unlock my-stack 9f8e7d6c-0000-0000-0000-000000000000
```

#### References

- [CDKTF CLI reference](https://developer.hashicorp.com/terraform/cdktf/cli-reference/commands)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/hashicorp/cdktf/README.md) — its plugin snippet calls `helm.NewCommand`, so ignore it
