#### Hazards

**Extracting overwrites files in the archive's own directory.** `unzip` runs with
the working directory set to the folder *containing* the archive, so it unpacks
alongside itself rather than into a subdirectory or the current directory - and
there is no flag to redirect it. Whatever paths the archive holds are written
relative to there, replacing existing files of the same name. Output is captured
rather than streamed, so `unzip`'s usual overwrite prompt cannot be answered.
Inspect an untrusted archive before extracting: nothing here constrains where its
entries land.

**`--cred` puts a plaintext password on the `unzip` command line.** The named
credential is fetched from 1Password and passed as `-P <password>`, visible in
the process list for the lifetime of the extraction. That is a property of
`unzip -P` rather than something this provider can avoid, but a password used
this way should be treated as exposed on any shared machine.

**`--cred` resolves a live 1Password secret**, so it needs a signed-in session
and may prompt. Without the flag `unzip` runs unauthenticated - for a
password-protected archive it then prompts on stdin and blocks.

#### Behaviour

A successful extraction prints nothing - no file list, no confirmation. Output is
surfaced only by wrapping it into a returned error, so silence means success.

An unknown `--cred` name fails with `credential <name> not found` before `unzip`
runs; suggestions come from the `credentials` keys in the posh config.

`extract` is the only verb here. The provider also exposes `Create` and
`CreateWithPassword` as a Go API for other providers - `postgres` uses them to
compress dumps - reachable from a plugin but never from the shell.

Completion finds `*.zip` anywhere in the project, skipping `vendor`,
`node_modules` and dot-directories.

#### Configuration

The single `zip` key `credentials` maps a name to a 1Password secret reference
(`account`/`vault`/`item`/`field`); the name is what `--cred` accepts and the
referenced field is used as the archive password. The config-key option lives on
`zip.New`, not on `NewCommand`, and this command cannot be renamed at all.

Field shapes: [`arbitrary/zip/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/arbitrary/zip/config.schema.json).

A `onepassword` provider must be wired into `zip.New`.

#### Examples

`<archive>` below stands for the path to the archive itself, `*.zip` extension
included - completion fills it in.

```bash
posh execute {{cmd}} extract <archive>
posh execute {{cmd}} extract <archive> --cred default
```

#### References

- [`unzip(1)`](https://linux.die.net/man/1/unzip) - the binary this wraps, and the `-P` caveat
- [Provider README](https://github.com/foomo/posh-providers/blob/main/arbitrary/zip/README.md) - config sample only
