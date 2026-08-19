#### Hazards

**Extracting overwrites files in the archive's own directory.** The command runs
`unzip` with the working directory set to the folder *containing* the zip, so
the archive unpacks alongside itself rather than into a subdirectory or the
current directory. Whatever paths the archive holds are written relative to
there, silently replacing existing files of the same name - `unzip` is run
non-interactively, so its usual overwrite prompt cannot be answered and the
extraction may fail or clobber depending on the archive. Inspect an untrusted
archive before extracting it: nothing here constrains where its entries land.

**`--cred` puts a plaintext password on the `unzip` command line.** The named
credential is fetched from 1Password and passed as `-P <password>`, so it is
visible in the process list for the lifetime of the extraction and may reach
shell history or process accounting. That is a property of `unzip -P`, not
something this provider can avoid, but it means an archive password should be
treated as exposed on any shared machine.

**Using `--cred` resolves a live 1Password secret**, so the command needs a
signed-in session and may trigger an interactive prompt. Without the flag no
secret is touched and `unzip` runs unauthenticated - which for a
password-protected archive means it prompts for a password on stdin and blocks.

#### Behaviour

Only `extract` exists in this tree. The underlying provider also implements
`Create` and `CreateWithPassword`, but neither is wired to a command - archives
can only be created by calling the Go API from a plugin, not from the shell.

The path you give is split: the command `cd`s to its directory and passes only
the basename to `unzip`. A relative path therefore behaves as expected, but the
extraction location always follows the archive, never your current directory,
and there is no flag to redirect it.

Output is captured rather than streamed, and surfaced only by wrapping it into
the returned error. A successful extraction prints nothing - no file list, no
confirmation - so silence means success here.

An unknown `--cred` name fails with `credential <name> not found` before `unzip`
runs. The suggestions come from the `credentials` keys in the posh config.

The `filename` completion finds `*.zip` anywhere in the project, skipping
`vendor`, `node_modules` and dot-directories.

#### Configuration

Config key `zip` by default, overridable via `WithConfigKey` on `zip.New` - note
this is an option on the provider constructor, not on `NewCommand`, and the
command itself cannot be renamed. Confirm the key against the project's own posh
config.

Field shapes: [`arbitrary/zip/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/arbitrary/zip/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

The single key `credentials` maps a name to a 1Password secret reference
(`account`/`vault`/`item`/`field`) - the name is what `--cred` accepts, and the
referenced field is used as the archive password. The provider requires a
`onepassword` provider to be wired into `zip.New`.

#### Examples

```bash
# Extracts into the directory containing the archive
posh execute zip extract path/to/archive.zip

# Password fetched from 1Password, then passed to unzip as -P
posh execute zip extract path/to/secret.zip --cred default
```

#### References

- [`unzip(1)`](https://linux.die.net/man/1/unzip) - the binary this wraps, and the `-P` caveat
- [Provider README](https://github.com/foomo/posh-providers/blob/main/arbitrary/zip/README.md) - config sample only
