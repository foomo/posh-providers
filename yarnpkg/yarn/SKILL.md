#### Hazards

**`run`'s `[path]` is declared optional but is read by index, so omitting it
runs the wrong thing.** Both arguments are positional, so `{{cmd}} run build`
sets the directory to `build` and leaves the script empty. Usually that fails on
a missing directory; if a directory of that name exists (`build/`, `test/` are
plausible) it silently runs an empty script in the wrong place. Always write
`{{cmd}} run <path> <script>`, using `.` for the root. The sibling `install`
gates its identically-named `[path]` on the argument count and genuinely
defaults to `.`, so the two behave differently in one tree.

**The root is a passthrough, so the three subcommands are not the real surface.**
Any unmatched argument goes straight to the upstream binary - `add`, `upgrade`,
`publish` and the rest - inheriting its side effects with nothing from posh in
front. Passthrough invocations always run in the **project root**; no `--cwd` is
added, so they ignore the `[path]` convention the subcommands use.

**`run-all` runs the script in every discovered package concurrently**, sharing
one errgroup context, so the first failure cancels the rest and leaves some
packages done, some half-done, some never attempted. It skips the root package
and takes no `[path]`, so there is no way to narrow it.

**`install` executes dependency lifecycle scripts** (arbitrary third-party code)
and writes a lockfile. There is no read-only verb in this tree.

#### Behaviour

**Discovery ignores more than it looks like.** The ignore patterns `^\.`, `dist`
and `node_modules` are matched **unanchored against each directory name** - so
`dist/` is skipped as intended, but so are `distribution/` and `my-dist-tools/`
and everything beneath them. A package whose directory name merely contains
"dist" is invisible to both completion and `run-all`. Cached per session.

**Only the passthrough root forwards flags.** The subcommands forward additional
args only, so a flag typed on one is silently dropped rather than reaching the
binary. Put such flags after `--`, which is forwarded, including the `--` token
itself. `run-all`'s `--parallel` is consumed by posh and never forwarded.

#### Configuration

No config key, no schema. Everything comes from the project's `package.json`
files and the tool's own configuration.

#### References

- [Upstream CLI reference](https://yarnpkg.com/cli) - the commands the root forwards to
