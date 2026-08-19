#### Hazards

`install` modifies the **system trust store**, and on Linux and macOS that
usually means a `sudo` password prompt from `mkcert` itself — so it blocks
waiting for a TTY rather than failing. Once it succeeds, every browser and tool
on the host trusts certificates minted by this CA until `uninstall` runs. Require
manual approval; it changes machine-wide state, not project state.

`install` also mints a stray certificate as a side effect. It runs
`mkcert -install install`, appending its own node name to the command line
because it forwards `r.Args()` unfiltered, and upstream treats a positional
argument after `-install` as a hostname to issue for. `install` matches
mkcert's hostname pattern, so nothing errors: the CA is installed and then
`install.pem` and `install-key.pem` are written **into the current working
directory** — the project root, since this verb sets no working directory —
overwriting those two files if they already exist. Delete them afterwards, and do
not read their presence as a configured certificate. `caroot` and `uninstall`
forward their node name the same way, but upstream returns before reading
positional arguments, so for those two it is inert.

`uninstall` removes the CA from the trust store without deleting it from disk, so
previously issued certificates stop being trusted while the files stay behind and
still look valid. Certificates already generated are not reissued by a later
`install`.

`generate` and `create` write private keys to the configured certificate path and
overwrite whatever is already there — mkcert performs no existence check and
prompts for nothing. `generate` rewrites *every* certificate in the config, not
just a changed one. Confirm the certificate path is git-ignored before running
either; the keys are unencrypted.

`caroot` is the only read-only verb, and it is the safe one for orientation: it
prints the CA storage location and issues nothing.

#### Behaviour

`generate` names its output from each entry's `name` field
(`<name>.pem`, `<name>-key.pem`) and passes that entry's `names` list as the
hostnames, so the filename and the certificate's validity are configured
separately and need not match. `create` lets mkcert choose the filename instead,
which derives it from the **first** name given, replacing `*` with `_wildcard`
and `:` with `_`. So `create '*.foomo.org' foomo.org` writes
`_wildcard.foomo.org.pem`, and the two verbs produce different filenames for the
same certificate.

`generate` and `create` run with the certificate path as their working directory,
so relative paths in flags resolve against it rather than the project root. Both
call the path's creation first and fail with `invalid empty path` when
`certificatePath` is unset, which is why an unconfigured project cannot
accidentally scatter keys into the checkout — unlike `install`, above.

`generate` stops at the first certificate that fails, leaving earlier entries
written and later ones untouched, so a partial run is not visible in the exit
status alone.

The tree declares no flags, but every verb forwards `r.Flags()` and
`r.AdditionalArgs()` verbatim to the binary, so upstream options such as
`-client`, `-ecdsa`, `-pkcs12` and `-csr` are reachable by typing them even
though nothing completes them. The root carries no `Execute`, so this is not a
passthrough: an unrecognised *subcommand* is rejected as an invalid command, and
a bare `mkcert` is rejected for a missing one. Only flags reach the binary
unlisted, and only through these five verbs.

#### Configuration

Read under the `mkcert` key by default; `WithConfigKey` renames it and
`CommandWithName` renames the command, so confirm against the project's own
config file. Field shapes:
[`filosottile/mkcert/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/filosottile/mkcert/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`certificates` is what `generate` iterates, so it — not this tree — defines which
certificates a project expects to exist. Read it to learn the expected filenames
before generating anything.

#### Examples

```bash
# Orientation: where the CA lives. Read-only.
posh execute mkcert caroot

# Trust the local CA. Prompts for sudo and writes a stray install.pem
# into the project root; needs approval.
posh execute mkcert install

# Regenerate every certificate listed in the config.
posh execute mkcert generate

# One-off certificate; the filename comes from the first name.
posh execute mkcert create foomo.org '*.foomo.org' localhost 127.0.0.1
```

#### References

- [mkcert README](https://github.com/FiloSottile/mkcert#readme)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/filosottile/mkcert/README.md) — titled "POSH doctl provider"; the plugin snippet and config sample are mkcert's
