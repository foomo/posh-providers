#### Hazards

**Omitting the path runs gotsrpc against every `gotsrpc.yml` in the project.**
The argument is optional, and when absent the provider walks the working
directory for every `gotsrpc.yml` and processes each in turn. So a bare
`gotsrpc` is the "regenerate everything" case, not a help screen.

**This command writes generated source files.** Each `gotsrpc.yml` decides which
Go and TypeScript files are produced and where, so running it overwrites
whatever is at those paths - including hand-edits made to generated files. The
command line says nothing about which files: read the `gotsrpc.yml` you are
about to run. The loop stops at the first failure, so a multi-file run can leave
some targets regenerated and others stale.

**Failures are only visible when the command fails.** Output is captured rather
than streamed, and surfaced only by wrapping it into the returned error. A run
that succeeds prints nothing but the per-path `gotsrpc: <path>` line, so
`--debug` output is invisible unless the run also errors.

#### Behaviour

`--debug` is the only flag this tree models, but any flag you pass is forwarded
after being rewritten: a leading `--` becomes a single `-`, because the upstream
binary uses single-dash long flags. Only the prefix is rewritten, so a value
containing `--` survives intact.

Validation is stricter than the argument name suggests: the path must be an
existing **file**, not a directory, and only one may be given. Both a missing
file and a directory produce `invalid [path] parameter`; two paths produce
`too many arguments`. Passing no path at all is valid and means "all".

The list of `gotsrpc.yml` files is discovered once and cached for the session,
so a file added after the shell started is not offered in completion and is not
included in a bare run until the cache is cleared.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type, no
`config.schema.json` and no `config.base.json`, so there is nothing to set in
the project's posh config and no entry for it in `posh.schema.json`. That is
deliberate, not an oversight: everything it needs comes from the `gotsrpc.yml`
files it discovers on disk.

Those files are the real configuration surface. Each one determines the packages
scanned and the Go/TypeScript output paths, so reading the target
`gotsrpc.yml` is the only way to know what a run will write.

The command takes no functional options - not even a rename - so it is always
registered as `gotsrpc`.

#### Examples

```bash
# Regenerate from one config file
posh execute gotsrpc path/to/gotsrpc.yml

# No path: regenerates from EVERY gotsrpc.yml in the project
posh execute gotsrpc
```

#### References

- [gotsrpc](https://github.com/foomo/gotsrpc) - the generator this wraps, and the `gotsrpc.yml` format
- [Provider README](https://github.com/foomo/posh-providers/blob/main/foomo/gotsrpc/README.md) - plugin wiring; note its snippet omits the required cache argument
