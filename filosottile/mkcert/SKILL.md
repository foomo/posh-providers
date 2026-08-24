#### Hazards

`install` modifies the **system trust store**, and on Linux and macOS that usually means
a `sudo` password prompt from the binary itself — so it blocks waiting for a TTY rather
than failing. Once it succeeds, every browser and tool on the host trusts certificates
minted by this CA until `uninstall` runs. Require manual approval: it changes
machine-wide state, not project state.

`uninstall` removes the CA from the trust store without deleting it from disk, so
previously issued certificates stop being trusted while the files stay behind and still
look valid.

`generate` and `create` write **unencrypted private keys** to the configured certificate
path and overwrite whatever is there — no existence check, no prompt. `generate` rewrites
*every* certificate in the config, not just a changed one. Confirm the certificate path is
git-ignored first.

`caroot` is the only read-only verb and the safe one for orientation: it prints the CA
storage location and issues nothing.

#### Behaviour

`generate` names its output from each entry's `name` field and passes that entry's `names`
list as the hostnames, so the filename and the certificate's validity are configured
separately and need not match. `create` lets the binary choose the filename instead,
deriving it from the **first** name given, replacing `*` with `_wildcard`. So the two verbs
produce different filenames for the same certificate.

Both run with the certificate path as their working directory, so relative paths in flags
resolve against it rather than the project root, and both fail with `invalid empty path`
when `certificatePath` is unset.

The tree declares no flags, but every verb forwards flags verbatim, so upstream options
such as `-client`, `-ecdsa` and `-pkcs12` are reachable by typing them even though nothing
completes them.

#### Configuration

Read under the `mkcert` key. `certificates` is what `generate` iterates, so it — not this
tree — defines which certificates a project expects to exist.

Field shapes: [`filosottile/mkcert/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/filosottile/mkcert/config.schema.json).

#### Examples

```bash
# Orientation: where the CA lives. Read-only.
posh execute {{cmd}} caroot

# Trust the local CA. Prompts for sudo; needs approval.
posh execute {{cmd}} install

# One-off certificate; the filename comes from the first name.
posh execute {{cmd}} create foomo.org '*.foomo.org' localhost
```

#### References

- [Upstream README](https://github.com/FiloSottile/mkcert#readme)
- `filosottile/mkcert/README.md` — titled "POSH doctl provider"; its snippet and config are this tool's
