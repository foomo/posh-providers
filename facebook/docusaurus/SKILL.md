#### Hazards

**Do not run this unattended - it never returns.** It builds a Docker image and
then starts the Docusaurus dev server in the foreground with `docker run -it`,
so it blocks until the container is killed, and it wants a TTY an agent
generally does not have. There is no flag to background it and no `stop` verb;
the way out is interrupting the process or stopping the container by hand.

`docker build` writes an image tagged `imageName:imageTag` from config, so
running this **overwrites any existing image with that tag** - config decides
what gets clobbered, not an argument. The Dockerfile is embedded in the provider
binary and piped in with `-f -`, so it is not a file in the checkout you can
inspect; the build context is the current working directory.

`docker run` publishes container port 3000 on the configured host port, so a
second invocation fails while the first is still running.

#### Behaviour

It takes no arguments, flags or subcommands, and nothing typed after the verb is
forwarded to the upstream CLI or to `npm` - there is no way to reach an upstream
flag from here. The whole invocation is assembled from config.

Prerequisites are external and unchecked: a working Docker daemon, and a site at
the configured source path containing a `package.json` (the build runs
`npm install` against it, then copies the rest). A wrong source path surfaces as
a Docker build failure, not a posh error.

#### Configuration

Read from the `docusaurus` key. Every part of the build and run comes from it,
so reading it is the only way to know what gets built, which image tag is
overwritten and which host port is bound.

Field shapes: [`facebook/docusaurus/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/facebook/docusaurus/config.schema.json).

Two things the schema cannot say: `volumes` entries pass through `os.ExpandEnv`,
so `$HOME`-style references resolve from the posh process's environment; and
`localPort` is only the host side - the container side is always 3000.

#### References

- [Docusaurus documentation](https://docusaurus.io/docs)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/facebook/docusaurus/README.md)
