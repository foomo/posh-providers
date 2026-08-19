#### Hazards

Most of this tree writes to remote storage, and the destructive verbs are named
after their upstream counterparts rather than after what they destroy. `purge`
removes a path **and all of its contents** where `delete` only removes files;
`deletefile`, `rmdir` and `rmdirs` remove single files and directories; `move`
and `moveto` delete the source once the copy lands. All of them act on live
object storage with no local staging and nothing to undo. Require manual
approval, and confirm the target remote before running any of them.

`sync` is the one to treat as the most dangerous. It makes destination identical
to source by **deleting whatever exists only at the destination**, and which side
that is depends entirely on the argument order — `sync a: b:` and `sync b: a:`
differ by one word and destroy opposite ends. Never run it unattended, and never
infer the direction from the shell history.

`init` overwrites the configured config file unconditionally — `os.WriteFile`
with no existence check — so running it discards any edits made to that file by
hand. It also shells out to `op inject`, so it needs a signed-in 1Password CLI
session; without one it fails, and the existing file is left alone only because
the write never happens. The credentials it materialises are written to disk
inside the project checkout, so confirm the path is git-ignored.

`mount` and `nfsmount` block for as long as the mount lives and need FUSE on the
host — they do not return, so an agent invoking either will hang rather than get
output. `dedupe` prompts for a keep/rename/delete decision per duplicate and
cannot be answered non-interactively. Prefer `lsf`, `ls`, `lsl`, `lsd`, `tree`,
`size`, `check`, `checksum` and `cat` for orientation: those are the read-only
verbs, and they are safe to run unattended against a remote whose contents are
not yet known.

The root carries its own `Execute` that forwards unmatched arguments straight to
the `rclone` binary, so the 26 verbs listed above are a hand-picked mirror, not
the real surface. Everything upstream supports is reachable and none of it is
listed here — including `config` (which rewrites the same file `init` owns),
`bisync`, `serve` and `ncdu`. A verb absent from the tree is not a verb the
provider blocks.

#### Behaviour

`<remote>` completes from `rclone listremotes`, which reads the generated config
file, so before `init` has run there are no remotes to offer. The suggester
splits that output on newlines without discarding empties, so an empty result
yields **one blank suggestion** rather than none — completion looks like a remote
exists when the config is missing entirely. It also discards the exec error
(`out, _ :=`), so a missing binary, an unreadable config and a genuinely empty
config are indistinguishable. Read the config key below to learn which remotes
should exist rather than trusting completion. The list is cached per shell
session, so remotes added by `init` or by upstream `rclone config` need a
`cache clear` before they appear.

Completions carry the trailing colon (`cloudflare:`) because the argument is
really rclone's `remote:path` operand. A bare local path works too, which is how
`copy`, `sync` and `move` address the local side; nothing distinguishes the two
in completion.

`RCLONE_CONFIG` is exported once, when the command is constructed at shell
startup, and points at the configured path whether or not a file is there. So
every rclone invocation in the session — including one typed through the
passthrough root — reads the project's config rather than the user's own
`~/.config/rclone`, and a `Validate` that only checks for the `rclone` binary
will not catch the file being absent.

Arguments are joined into a single `sh -c` line, so the resolved secret values
are subject to shell word splitting and metacharacter expansion. That matters
for the config template rather than for the arguments: a secret containing shell
metacharacters can break `init` or leak into the rendered file mangled.

#### Configuration

Read under the `rclone` key by default; `WithConfigKey` renames it, so confirm
against the project's own config file. Field shapes:
[`rclone/rclone/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/rclone/rclone/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`config` holds the whole rclone INI file as a string, so it — not this tree — is
where the set of available remotes is defined. Read it to learn which remotes
exist and what backend each one points at.

#### Examples

```bash
# Orientation: what remotes are configured, then what is in one.
posh execute rclone lsd cloudflare:
posh execute rclone lsl cloudflare:bucket/dir

# Compare without writing anything.
posh execute rclone check ./local cloudflare:bucket

# Upload without deleting at the destination — prefer this over sync.
posh execute rclone copy ./local cloudflare:bucket/dir

# Materialise the config file; needs a signed-in 1Password CLI.
posh execute rclone init
```

#### References

- [rclone commands](https://rclone.org/commands/)
- [rclone config file format](https://rclone.org/docs/#config-config-file)
- [`op inject`](https://developer.1password.com/docs/cli/reference/commands/inject)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/rclone/rclone/README.md)
