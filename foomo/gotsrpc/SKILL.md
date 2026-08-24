#### Hazards

**Omitting the path runs the generator against every `gotsrpc.yml` in the
project.** The argument is optional, and when absent the provider walks the
working directory and processes each config in turn. A bare invocation is the
"regenerate everything" case, not a help screen.

**This command writes generated source files.** Each `gotsrpc.yml` decides which
Go and TypeScript files are produced and where, so a run overwrites whatever is
at those paths - including hand-edits made to generated files. The command line
says nothing about which files: read the `gotsrpc.yml` first. The loop stops at
the first failure, so a multi-config run can leave some targets regenerated and
others stale.

**Output is only visible when the command fails.** It is captured rather than
streamed and surfaced solely by wrapping it into the returned error. A
successful run prints nothing but one line per path, so `--debug` output is
invisible unless the run also errors.

#### Behaviour

`--debug` is the only flag this tree models, but any flag you pass is forwarded
after being rewritten: a leading `--` becomes a single `-`, because the upstream
binary uses single-dash long flags. Only the prefix is rewritten, so a value
containing `--` survives intact.

The path argument must be an existing **file**, not a directory, and only one
may be given. Passing none is valid and means "all".

The config list is discovered once and cached for the session, so a file added
after the shell started is neither offered in completion nor included in a bare
run.

#### Configuration

**This provider has no config key of its own** - no `Config` type, no
`config.schema.json`, nothing to set in the project's posh config. Everything it
needs comes from the `gotsrpc.yml` files it discovers on disk.

Those files are the real configuration surface: each determines the packages
scanned and the Go/TypeScript output paths, so reading the target one is the
only way to know what a run will write.

The command takes no functional options, not even a rename.

#### Examples

Pass the path of one `gotsrpc.yml` to scope the run to it, or omit the argument
to regenerate from every one in the project.

```bash
posh execute {{cmd}}
```

#### References

- https://github.com/foomo/gotsrpc for the generator and the `gotsrpc.yml` format
- `foomo/gotsrpc/README.md` for plugin wiring
