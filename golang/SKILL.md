#### Hazards

Most verbs are read-only or write only build artefacts. The exceptions:

`mod upgrade` rewrites `go.mod`/`go.sum`, bumping every direct dependency with an
available update to `@latest` - no confirmation, no dry run. With `[path]`
omitted it does this to **every** module in the checkout at once.

`clean mod` deletes the **shared, machine-wide** module download cache
(`-modcache`), so every other Go project on the host re-downloads afterwards.
The sibling caches are machine-wide too but only cost rebuild time.

`generate` executes arbitrary code, defined by the project's `generate.go`
directives rather than by this provider - read them first. With `[path]` omitted
it runs every one in the checkout.

`fuzz` **with a target named** fuzzes, running until it finds a failure or is
interrupted, so it must not be run unattended. Without a target it replays the
existing corpus once and terminates. The distinction is one argument.

`lint --fix` rewrites source files in place, and **`lint` enrols this provider in
`arbitrary/lint`**: that aggregator discovers it by type assertion, so a
project-level lint sweep runs `golangci-lint` across every `go.mod` here and
forwards `fix` - the in-place rewrite then happens as part of a command that
never mentions Go. That entry point ignores `--parallel` and `[path]`; it always
walks every module, serially.

#### Behaviour

**Omitting `[path]` widens to every module**, each verb falling back to walking
for `go.mod`. The usage block shows it as optional, which reads as less work
rather than more.

**Positional arguments are read by index**, so with `[path] [package] [target]`
on `test`/`fuzz`/`bench`, passing only `./pkg/x` is read as the *module* path.

**Build tags come from `GO_BUILD_TAGS` (default `safe`), not from `--tags`**,
which only sets a `GO_TEST_TAGS` variable that nothing here reads, so it is
inert.

**Anything after `--` replaces the build tags rather than adding to them**, so
`-- -count=1` compiles without `-tags safe` and silently builds a different set
of files. Pass `-tags safe` explicitly alongside anything else after `--`.

**`--parallel N` shares one errgroup context**, so the first failure cancels the
rest mid-run; the default is serial. Completion shells out, so completing an
argument compiles code.

`work sync` is the idempotent form - `work init` aborts against an existing
`go.work` - and it *adds* every module that appeared under `.`.

#### Configuration

No config key, no schema; the only external input is `GO_BUILD_TAGS`.

#### References

- [Upstream toolchain reference](https://pkg.go.dev/cmd/go)
- [golangci-lint](https://golangci-lint.run/) - used by `lint` and `clean lint`
