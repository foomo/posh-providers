#### Hazards

**Do not run this unattended - it never returns.** The command builds a Docker
image and then starts the Docusaurus dev server in the foreground with
`docker run -it`, so it blocks until the container is killed. It also requires a
TTY (`-it`), which an agent generally does not have. There is no flag to
background it and no separate "stop" verb; the way out is interrupting the
process or stopping the container by hand.

Despite taking no arguments, it is **not** a thin wrapper: the whole invocation is
assembled from the config key, and unmatched arguments are *not* forwarded to
`docusaurus` or `npm`. Whatever you type after the command name is ignored, so
there is no way to reach an upstream flag from here.

Two side effects outside the checkout, both from Docker rather than posh:

`docker build` writes an image tagged from the config (`imageName:imageTag`), so
running this **overwrites any existing image with that tag** - the config decides
what gets clobbered, not an argument. The Dockerfile is embedded in the provider
binary and piped in via `-f -`, so it is not a file in the checkout you can
inspect or edit; the build context is the current working directory.

`docker run` publishes the container's port 3000 on the configured host port, so
the command fails if that port is already bound - which is what a second, still
running invocation looks like.

Prerequisites are external and unchecked: a working Docker daemon, and a
Docusaurus site at the configured source path containing a `package.json` (the
build does `npm install` against it, then copies the rest). A missing or wrong
source path surfaces as a Docker build failure, not a posh error.

#### Configuration

Config key `docusaurus` by default, renameable via `CommandWithConfigKey` (and the
command itself via `CommandWithName`), so confirm both against the project's own
posh config. Every part of the build and run is taken from this key - there are no
arguments or flags - so reading it is the only way to know what will be built,
which image tag is overwritten, and which host port is bound.

Field shapes: [`facebook/docusaurus/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/facebook/docusaurus/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same
schema is bundled under in the project's `posh.schema.json`.

Two things the schema cannot say: `volumes` entries are passed through
`os.ExpandEnv`, so `$HOME`-style references in them are resolved from the posh
process's environment; and `localPort` is only the host side - the container side
is always 3000.

#### Examples

```bash
# Builds the image, then blocks serving the site on the configured localPort
posh execute docusaurus
```

#### References

- [Docusaurus documentation](https://docusaurus.io/docs)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/facebook/docusaurus/README.md)
