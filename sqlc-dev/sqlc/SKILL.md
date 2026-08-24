#### Hazards

**`generate` writes source files, overwriting what is there.** It regenerates the
output each `sqlc.yaml` declares, replacing existing files with no diff and no
confirmation - hand-edits to generated code are lost. The other three leaves
(`compile`, `diff`, `vet`) are read-only checks.

**Omitting the path runs the verb against every `sqlc.yaml` in the project.** The
argument is optional, and with none given the provider walks the working
directory and processes each config in turn, so a bare `generate` regenerates
everything rather than the package you were working in. The loop stops at the
first failure, leaving some outputs regenerated and others stale.

**The root is a passthrough**, so any upstream subcommand not listed here -
`init`, `createdb`, `push`, `verify`, `upload` - is still reachable by typing it
and runs with its full upstream effects, including ones that talk to the vendor's
cloud service. The four listed leaves are not the limit of what this command can
do.

#### Behaviour

`--no-remote` is passed on every invocation, including the passthrough, disabling
remote/cloud features regardless of what the config file asks for.

Each run sets the working directory to the folder containing the config file, so
relative paths inside it resolve as upstream expects.

The four leaves capture output rather than streaming it, surfacing it only by
wrapping it into an error, so a successful run prints nothing beyond the per-path
progress line. The root passthrough streams normally.

Additional args after a `--` separator are forwarded; additional *flags* are not,
and the leaves declare no flags, so upstream flags are only reachable through the
passthrough form or after `--`. Discovery matches the filename `sqlc.yaml` only -
a project using `sqlc.json` or `sqlc.yml` is invisible - and the list is cached
for the session.

#### Configuration

Under the `sqlc` key, both fields are directories resolved relative to the
project root and passed as environment variables: `cacheDir` as `SQLCCACHE`,
`tempDir` as `SQLCTMPDIR`. Leaving either unset resolves that variable to the
project root itself, so cache or temp files land in the checkout.

Field shapes: [`sqlc-dev/sqlc/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/sqlc-dev/sqlc/config.schema.json).

What each config *generates* is decided by the `sqlc.yaml` files themselves -
reading the target file is the only way to know what `generate` will write.

#### Examples

```bash
posh execute {{cmd}} vet <path-to-a-config-file>
posh execute {{cmd}} generate
```

#### References

- [sqlc documentation](https://docs.sqlc.dev/) - the real command surface and the `sqlc.yaml` format
- [Provider README](https://github.com/foomo/posh-providers/blob/main/sqlc-dev/sqlc/README.md)
