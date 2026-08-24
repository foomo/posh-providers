#### Hazards

**This command deletes files before generating.** For each config it runs a
`rm -f` over the `gocontentful*.go` glob in the directory holding the
`gocontentful.yaml`, then regenerates. A shell expands that glob, so anything
matching is removed whether or not it was generated - a hand-written
`gocontentful_helpers.go` goes with it, no confirmation and no backup. The
deletion happens even when generation then fails, leaving no client code at all.

**Omitting the path does this for every `gocontentful.yaml` in the project.**
With no argument the provider walks the working directory and
deletes-and-regenerates in each config's directory in turn, stopping at the
first failure - so earlier directories are rewritten and later ones untouched.

**Every run resolves 1Password secrets and reaches the Contentful API.** The
config is rendered through the onepassword provider first, so a `cmaKey` held as
a secret reference is fetched on each invocation - needing a signed-in session
and possibly a prompt. The resolved management API key is handed to the upstream
binary, which talks to Contentful with it. This is not an offline generator.

**Config values are interpolated into a shell command line.** posh runs the
joined arguments through `sh -c`, so the resolved key is subject to shell
interpretation and appears in the process list. Treat a `gocontentful.yaml` as
trusted input.

#### Behaviour

`environment` defaults to `master` when empty, and that substitution happens in
the provider, not in the binary.

The path argument must be an existing **file**, not a directory.

Discovery matches `gocontentful.yaml` exactly, skipping `node_modules` and
dot-directories, and is cached for the session. A project using
`gocontentful.yml` is invisible to it.

`--debug` is declared but never forwarded; only args after a `--` separator
reach the binary.

#### Configuration

**No posh config key is read.** The provider ships a `config.base.json` that
registers `gocontentful` in the project's `posh.schema.json`, but `viper` is
never consulted, so setting that block has no effect. The schema at
[`foomo/gocontentful/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/gocontentful/config.schema.json)
documents the shape of each `gocontentful.yaml` instead. Reading the target file
is the only way to know which space, environment and content types a run
touches.

The `onepassword` provider must be wired in.

#### Examples

Pass one `gocontentful.yaml` path to scope the run; omit it to regenerate
everywhere.

```bash
posh execute {{cmd}}
```

#### References

- https://github.com/foomo/gocontentful
- `foomo/gocontentful/README.md`; its "Config" sample is a `gocontentful.yaml`, not a posh config key
