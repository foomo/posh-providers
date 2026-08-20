#### Hazards

Most verbs are read-only or write only build artefacts. Four mutate the checkout or the machine and
need a human to approve them, and one does not terminate on its own:

`mod upgrade` rewrites `go.mod`/`go.sum`. It pipes `go list -u -m` into
`xargs go get {}@latest`, so every direct dependency with an available update is bumped to its
latest version — there is no confirmation step and no dry run. With `[path]` omitted it does this to
**every** module in the checkout at once. Review the resulting diff before anything else runs.

`clean mod` deletes the **shared, machine-wide** module download cache (`go clean -modcache`), not
anything scoped to this project. Every other Go project on the host re-downloads its dependencies
afterwards. The three sibling caches (`build`, `test`, `fuzz`) are also machine-wide but only cost
rebuild time.

`lint --fix` rewrites source files in place. See also the note below on `lint` being enrolled in
`arbitrary/lint`.

`fuzz` with a target named runs `go test -fuzz`, which runs **until it finds a failure or is
interrupted** — it does not terminate on its own, so it must not be run unattended. Without a target
it degrades to `-run ^Fuzz`, which replays the existing corpus once and terminates; that form is
safe. The distinction is one argument.

**Running `lint` enrols this provider in `arbitrary/lint`.** It implements `Lint(ctx, fix)`, which
`arbitrary/lint` discovers by type assertion over the command registry, so `lint` at the project
level runs `golangci-lint run` across every `go.mod` here, and `lint --fix` forwards `fix` — the
in-place rewrite above then happens as part of a command that never mentions Go. `Lint` ignores the
`--parallel` flag and the `[path]` argument entirely; it always walks every module, serially.

`generate` executes arbitrary code. It runs `go generate` against the project's `generate.go` files,
so what it does — usually writing generated sources over existing ones — is defined by those
directives, not by this provider. Read them before running it, and note that with `[path]` omitted it
runs every `generate.go` in the checkout.

Everything else leaves the checkout alone: `mod outdated`, `test`, `bench` and `build` (output goes
to `/dev/null`, so it only type-checks and links). `mod download` and the completion helpers populate
the module cache but write no project files.

#### Behaviour

**Omitting `[path]` widens to every module.** Each verb falls back to
`files.Find(ctx, ".", "go.mod")`, ignoring `node_modules` and dot-directories, and runs against all
of them. The usage block shows `[path]` as optional, which reads as less work rather than more.
`generate` does the same with `generate.go` — and passes the *file* path to `go generate`, whereas
every other verb passes the *containing directory* as the working directory. Both lists are cached
per session, so a newly added module needs a `cache clear`.

**Positional arguments are order-dependent and there is no way to skip one.** `test`, `fuzz` and
`bench` take `[path] [package] [target]` and read them by index, so naming a package requires naming
the module directory first: `go test . ./pkg/x TestFoo`, not `go test ./pkg/x`. Passing only
`./pkg/x` is interpreted as the *module* path. `build` is the same shape with a repeatable trailing
`[package]`.

**Build tags come from `GO_BUILD_TAGS`, not from `--tags`.** Every `go` invocation gets
`-tags $GO_BUILD_TAGS` with `safe` as the default. The `--tags` flag on `test`/`fuzz`/`bench` sets a
`GO_TEST_TAGS` environment variable for the child process instead and never reaches the command
line, so it only has an effect if the project's own tooling reads it — nothing in this provider does.

**Anything after `--` replaces the build tags rather than adding to them.** `test`, `fuzz`, `bench`
and `build` assign `r.AdditionalArgs().From(1)` over the `-tags` slice, so `go test -- -count=1`
runs without `-tags safe` and silently compiles a different set of files. Pass `-tags safe`
explicitly alongside anything else after `--`.

**Default parallelism is 1.** `--parallel N` raises the errgroup limit; omitted, modules are
processed one at a time. The group shares one context, so the first failure cancels the rest
mid-run, leaving later modules unattempted. `lint` additionally adds `--allow-parallel-runners` to
golangci-lint whenever `--parallel` is set, so the two run concurrently without fighting over the
lint cache.

**Flags typed on `lint` reach golangci-lint re-rendered from the parsed flag set, not verbatim**, so
`--timeout 5m` arrives as `--timeout 5m0s`. Internal flags such as `--parallel` are consumed by posh
and never forwarded. Anything that golangci-lint accepts but this tree does not declare has to go
after `--`.

**`work init` and `work sync` are both three-and-two-step sequences, not single commands.** `init`
runs `go work init`, `go work use -r .`, `go work sync`, aborting on the first failure — so against an
existing `go.work` it stops at `go: go.work already exists` and changes nothing. `sync` runs the last
two, which makes it the idempotent form and means it also *adds* every module that appeared under `.`
since, not just synchronises the existing entries.

**`work use` forwards the whole argument vector**, producing `go work use <path>` — the node names
are the upstream subcommand names, which is why it works. It is the only verb that reaches `go`
without a `-tags` argument.

**Completion shells out.** `[package]` runs `go list` in the chosen module, and `[target]` runs
`go test -list` against the chosen package — so completing an argument compiles code, and the
suggestions are empty when the module does not build. `build`'s `[package]` lists only `main`
packages; `test`/`fuzz`/`bench` list only packages containing test files, rendered relative to their
module.

**`mod outdated` requires `go-mod-outdated` on `PATH`** — it pipes `go list -u -m -json all` into it
and fails with a shell error if the binary is absent. `mod upgrade` pipes through `xargs`. Both work
because posh's `shell` helper joins its arguments into a single `sh -c` line, so these are real shell
pipes; by the same token every value the other verbs interpolate is subject to shell word splitting
and metacharacters.

#### Configuration

None. This provider reads no config key and ships no schema; the only external input is the
`GO_BUILD_TAGS` environment variable (default `safe`). `CommandWithExecGolangciLint` replaces how the
`golangci-lint` binary is invoked, which is how a project points at an ownbrew-managed copy.

#### Examples

```bash
# tidy every module in the checkout, four at a time
posh execute go mod tidy --parallel 4

# one module only
posh execute go mod tidy ./pkg/foo

# a single test in a single package - the module path is required first
posh execute go test . ./pkg/foo TestBar

# extra go test flags, keeping the build tag that -- would otherwise drop
posh execute go test . ./pkg/foo -- -tags safe -count=1 -race

# lint one module and fix in place (rewrites files)
posh execute go lint ./pkg/foo --fix
```

#### References

- [go command](https://pkg.go.dev/cmd/go) — the underlying subcommands
- [golangci-lint](https://golangci-lint.run/) — used by `lint` and `clean lint`
- [go-mod-outdated](https://github.com/psampaz/go-mod-outdated) — required by `mod outdated`
- [provider README](README.md)
