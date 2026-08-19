#### Hazards

**This command deletes files before generating.** For each config it runs
`rm -f <dir>/gocontentful*.go` in the directory containing the
`gocontentful.yaml`, then regenerates. Anything matching that glob is removed
whether or not it was generated - a hand-written `gocontentful_helpers.go`
beside the generated code is deleted with no confirmation and no backup. The
deletion happens even if generation subsequently fails, leaving the directory
empty of client code.

**Omitting the path does this for every `gocontentful.yaml` in the project.**
The argument is optional, and with none given the provider walks the working
directory and processes each config in turn - deleting and regenerating in each
one. The loop stops at the first failure, so earlier directories are already
rewritten and later ones untouched.

**Every run resolves 1Password secrets and reaches the Contentful API.** The
config file is passed through the onepassword provider's `RenderFile` before
being parsed, so a `cmaKey` held as a secret reference is fetched on each
invocation - requiring a signed-in session and possibly an interactive prompt.
The resolved management API key is then handed to the `gocontentful` binary,
which talks to Contentful with it. This is not an offline code generator.

**The CMA key is interpolated into a shell command line.** posh's shell helper
joins all arguments and runs them through `sh -c`, so the resolved key - and
every other config value - is subject to shell interpretation and appears in the
process list. A key or content type containing shell metacharacters would be
mis-parsed or worse. Treat the values in a `gocontentful.yaml` as trusted input.

#### Behaviour

**`Config` here is not a posh config key.** Despite the provider shipping a
`config.base.json` that registers `gocontentful` in the project's
`posh.schema.json`, nothing reads that key - `viper` is never consulted. The
struct describes the schema of the **`gocontentful.yaml` files discovered on
disk**, which is what the schema link below documents. Setting a `gocontentful:`
block in the posh config has no effect.

`environment` defaults to `master` when empty - Contentful's built-in default -
and that substitution happens in the provider, not in the binary.

Validation is stricter than the argument name suggests: the path must be an
existing **file**, not a directory, and only one may be given. Both a missing
file and a directory produce `invalid [path] parameter`.

Discovery matches `gocontentful.yaml` exactly, skipping `node_modules` and
dot-directories, and is cached for the session - a config added after the shell
started is not offered or processed until the cache is cleared. A project using
`gocontentful.yml` is invisible to it.

The `--debug` flag is declared but never forwarded; only additional args after a
`--` separator reach the binary.

#### Configuration

The provider takes no posh config of its own - see the hazard above. What is
documented at
[`foomo/gocontentful/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/gocontentful/config.schema.json)
is the shape of each `gocontentful.yaml`: `spaceId`, `cmaKey`, `environment`,
`region` and `contentTypes`. That URL is also the schema's `$id`, so it is the
`$defs` key the same schema is bundled under in the project's
`posh.schema.json` - where it is misleadingly attached to a `gocontentful` key
that is never read.

Reading the target `gocontentful.yaml` is the only way to know which space,
environment and content types a run will touch.

The command needs a `onepassword` provider wired into `NewCommand`, and can be
renamed with `CommandWithName`.

#### Examples

```bash
# Deletes gocontentful*.go beside this config, then regenerates
posh execute gocontentful path/to/gocontentful.yaml

# Same, for EVERY gocontentful.yaml in the project
posh execute gocontentful
```

#### References

- [gocontentful](https://github.com/foomo/gocontentful) - the generator this wraps
- [Provider README](https://github.com/foomo/posh-providers/blob/main/foomo/gocontentful/README.md) - note its "Config" sample is a `gocontentful.yaml`, not a posh config key
