#### Hazards

**The root is a passthrough, so the three subcommands are not the real surface.** Any unmatched
argument goes straight to the `yarn` binary — `yarn add`, `yarn upgrade`, `yarn publish`, `yarn cache
clean` and everything else upstream offers is reachable and none of it is listed above. The
passthrough inherits yarn's own side effects with nothing from posh in front: `yarn add` and
`yarn upgrade` rewrite `package.json` and the lockfile, `yarn publish` pushes a package to a registry.
Treat an unrecognised verb as running upstream unmodified, and check what it does before running it.
Passthrough invocations always run in the **project root** — no `--cwd` is added — so they ignore the
`[path]` convention the subcommands use.

**`run-all` runs the script in every discovered package concurrently.** With no `--parallel` the limit
is 1, so it is serial; with `--parallel N` up to N packages run at once, sharing one errgroup context,
so the first failure cancels the rest mid-run and leaves some packages done, some half-done and some
never attempted. What the script itself does is defined by each `package.json`, so the blast radius is
whatever the project's `build`/`deploy`/`release` scripts do — read them before running this across
the repo.

**`install` writes a lockfile and executes package lifecycle scripts.** `yarn install` runs
`preinstall`/`postinstall` hooks from every installed dependency, which is arbitrary third-party code.
Normal for a JS project, but it is not a read-only command.

There is no read-only verb in this tree. Everything either installs, runs a project script, or
forwards to yarn.

#### Behaviour

**`run`'s `[path]` is declared optional but is effectively required.** Both arguments are read by
index — `dir = At(1)`, `script = At(2)` — so `yarn run build` sets `dir` to `"build"` and leaves
`script` empty, running `yarn run ""` inside a directory named after the script. Usually that fails
with a missing-directory error; if a directory of that name happens to exist (`build/`, `test/` and
`lint/` are plausible), it silently runs the wrong thing in the wrong place. Script completion
confirms the intent: it reads `package.json` from `At(1)`, so nothing is suggested until a path is
typed first. Always write `yarn run <path> <script>`, using `.` for the root. Verified by running the
argument vector through `readline.Parse`.

**`install`'s `[path]` genuinely is optional** and falls back to `.` — it is gated on `LenGt(1)`
rather than read blindly, so the two nodes' identically-named argument behaves differently.

**`run-all` skips the root package.** Discovery finds `./package.json` like any other, but the loop
explicitly excludes `.`, so the root's own scripts are never run — only nested packages. `run-all` also
does not accept a `[path]`, so there is no way to narrow it.

**Discovery ignores more than it looks like.** Packages are found by walking for `package.json` with
the ignore patterns `^\.`, `dist` and `node_modules`, matched **unanchored against each directory
name**. So `dist/` is skipped as intended, but so are `distribution/`, `my-dist-tools/` and everything
beneath them — a real package whose directory name merely contains "dist" is invisible to both
completion and `run-all`. Verified by walking a tree containing each case. The list is cached per
session, so a newly added package needs a `cache clear`.

**Only the passthrough root forwards flags.** `execute` passes `r.Args()`, `r.Flags()` and
`r.AdditionalArgs()`, while `install`, `run` and `run-all` forward `AdditionalArgs()` only — so a flag
typed on a subcommand is silently dropped instead of reaching yarn: `yarn run ./web build --verbose`
runs without `--verbose`. Verified by parsing both forms. Put such flags after `--`, which is
forwarded — including the `--` token itself, so the command becomes
`yarn run build -- --verbose` (harmless, since yarn uses `--` as its own separator for passing
arguments to a script).

`run-all`'s `--parallel` is declared on the *default* flag set yet consumed by posh, so it is the one
flag that is read and never forwarded.

**The command can be renamed.** `CommandWithName` changes the command's own name (default `yarn`),
which is how a project exposes it as something else; the binary invoked is always `yarn`.

#### Configuration

None. This provider reads no config key and ships no schema. Everything comes from the project's
`package.json` files and yarn's own configuration.

#### Examples

```bash
# install in the root
posh execute yarn install

# install in one package
posh execute yarn install ./web

# run a script - the path is required, use . for the root
posh execute yarn run . build
posh execute yarn run ./web build

# extra flags must go after --, subcommands drop flags otherwise
posh execute yarn run ./web build -- --verbose

# run a script in every nested package, four at a time (root excluded)
posh execute yarn run-all build --parallel 4

# anything unmatched goes straight to yarn, in the project root
posh execute yarn add --dev typescript
```

#### References

- [yarn cli](https://yarnpkg.com/cli) — the upstream commands the root forwards to
- [provider README](README.md)
