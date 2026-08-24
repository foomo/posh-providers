#### Hazards

Most of this tree writes to remote storage, and the destructive verbs are named after
their upstream counterparts rather than after what they destroy. `purge` removes a path
**and all of its contents** where `delete` only removes files; `move` and `moveto` delete
the source once the copy lands. All act on live object storage with no local staging and
nothing to undo. Require manual approval and confirm the target remote first.

`sync` is the most dangerous: it makes destination identical to source by **deleting
whatever exists only at the destination**, and which side that is depends entirely on
argument order — reversing the operands destroys the opposite end. Never run it
unattended, and never infer the direction from shell history.

`init` overwrites the configured config file unconditionally, so running it discards hand
edits. It shells out to the 1Password CLI and needs a signed-in session. The credentials it
materialises land on disk inside the checkout — confirm that path is git-ignored.

`mount` and `nfsmount` block for as long as the mount lives and need FUSE, so an agent
invoking either hangs. `dedupe` prompts per duplicate and cannot be answered
non-interactively. Prefer the listing verbs plus `size`, `check`, `checksum` and `cat` for
orientation: those are the read-only ones.

**The root forwards unmatched arguments straight to the binary**, so the verbs listed are
a hand-picked mirror, not the real surface. Everything upstream supports is reachable and
unlisted — including the `config` verb, which rewrites the same file `init` owns, plus
`bisync`, `serve` and `ncdu`. A verb absent from the tree is not one the provider blocks.

#### Behaviour

`RCLONE_CONFIG` is exported once at shell startup and points at the configured path
whether or not a file is there. So every invocation in the session reads the project's
config rather than the user's own, and the `Validate` check — which only looks for the
binary — will not catch the file being absent.

#### Configuration

Read under the `rclone` key. `config` holds the whole INI file as a string, so it — not
this tree — defines which remotes exist and what backend each points at.

Field shapes: [`rclone/rclone/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/rclone/rclone/config.schema.json).

#### Examples

```bash
# Orientation, read-only.
posh execute {{cmd}} lsd cloudflare:

# Upload without deleting at the destination - prefer this over sync.
posh execute {{cmd}} copy ./local cloudflare:bucket/dir
```

#### References

- [Commands](https://rclone.org/commands/)
- [Config file format](https://rclone.org/docs/#config-config-file)
