#### Hazards

**`generate` writes source files, overwriting what is there.** It regenerates
the Go/TypeScript output each `sqlc.yaml` declares, replacing existing files
with no diff or confirmation - hand-edits to generated code are lost. The other
three leaves (`compile`, `diff`, `vet`) are read-only checks.

**Omitting the path runs the verb against every `sqlc.yaml` in the project.**
The argument is optional, and with none given the provider walks the working
directory and processes each config in turn. So a bare `sqlc generate`
regenerates everything, not just the package you were working in. The loop stops
at the first failure, leaving some outputs regenerated and others stale.

**The root is a passthrough**, so any `sqlc` subcommand not listed here -
`init`, `createdb`, `push`, `verify`, `upload` - is still reachable by typing it
and runs with whatever effects it has upstream, including ones that talk to
sqlc's cloud service. The four listed leaves are not the limit of what this
command can do.

#### Behaviour

`--no-remote` is passed on every invocation, including the passthrough, which
disables sqlc's remote/cloud features regardless of what the config file asks
for.

Each run happens with the working directory set to the folder containing the
`sqlc.yaml`, so relative paths inside that file resolve as sqlc expects.

The four leaves capture output rather than streaming it, and surface it only by
wrapping it into an error - so a successful `generate` or `vet` prints nothing
beyond the per-path progress line. The root passthrough streams normally.

Additional args after a `--` separator are forwarded; additional *flags* are
not, and the leaves declare no flags of their own - so upstream flags are only
reachable through the passthrough form or after `--`.

The list of `sqlc.yaml` files is discovered once and cached for the session, so
a config added after the shell started is not offered or processed until the
cache is cleared. Note discovery matches `sqlc.yaml` only - a project using
`sqlc.json` or `sqlc.yml` is invisible to it.

#### Configuration

Config key `sqlc` by default *in intent* - see the hazard above; overridable via
`CommandWithConfigKey` (and the command renameable via `CommandWithName`).

Field shapes: [`sqlc-dev/sqlc/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/sqlc-dev/sqlc/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

Both fields are directories resolved relative to the project root and passed to
sqlc as environment variables - `cacheDir` as `SQLCCACHE`, `tempDir` as
`SQLCTMPDIR`. Leaving either unset resolves that variable to the project root
itself, so sqlc writes its cache or temp files into the checkout.

What each config *generates* is decided by the `sqlc.yaml` files themselves, not
by anything here - reading the target `sqlc.yaml` is the only way to know what
`generate` will write.

#### Examples

```bash
# Read-only check of one config
posh execute sqlc vet path/to/sqlc.yaml

# Writes generated sources for EVERY sqlc.yaml in the project
posh execute sqlc generate

# Passthrough: any other sqlc subcommand
posh execute sqlc version
```

#### References

- [sqlc documentation](https://docs.sqlc.dev/) - the real command surface and the `sqlc.yaml` format
- [Provider README](https://github.com/foomo/posh-providers/blob/main/sqlc-dev/sqlc/README.md) - note its config sample has a `cacheDirDir` typo
