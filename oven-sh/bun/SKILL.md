#### Hazards

**Several verbs modify the project, and `run` executes arbitrary package
scripts.** `install`, `add`, `update` and `remove` rewrite `package.json` and
the lockfile and fetch from the network; `x` **downloads and executes a package
binary, installing it if it is not present**. `run <script>` executes whatever
that script is defined as in `package.json`, so its blast radius is that file's
contents, not anything visible here. Only `outdated` and a read-only `pm`
subcommand are safe unattended.

**The tree is not the surface.** The root and every leaf forward to the real
binary, so subcommands this tree does not list - `build`, `init`, `link`, `repl`
and the rest - are still reachable by typing them. Treat this as "all of the
toolkit", not the listed verbs.

**`workspace <path> <verb>` runs the verb with the working directory set to that
package**, so the same `install` or `remove` affects a different `package.json`
depending on one argument. The top-level forms run in the project root. Check
which of the two you are using before any mutating verb.

**`--watch` and `--hot` do not return.** Both keep the process running and
restart on file change, so a command carrying either blocks until interrupted.
`--inspect-wait` and `--inspect-brk` also block, waiting for a debugger.

#### Behaviour

Flags pass through **verbatim as typed**, not re-rendered from the parsed flag
set, so this tree's declarations only drive completion. That matters because
several value-taking upstream options are declared here as booleans -
`--install`, `--preload`, `--omit`, `--print` - so completion offers them as
bare switches, while `--omit=dev` typed in full is forwarded correctly. Trust
upstream documentation for a flag's shape, not this tree's suggestion. The list
is also partial and hand-maintained: options commented out in the source
(`--port`, `--config`, `--cwd`, `--silent`) are absent from completion yet still
work when typed.

**Nothing after a `--` separator is forwarded** - unusual among these providers,
so that escape hatch does not work here.

Workspace paths come from the root `package.json`'s `workspaces.packages`
(trailing `/*` stripped); with none declared the walk starts at `.`. Script names
come from the target directory. Both are cached per session.

#### Configuration

No config key, no schema. What is reachable is decided by the project's
`package.json` files; upstream behaviour by its own `bunfig.toml`. The command
cannot be renamed; the only option swaps the binary invocation.

#### References

- [Upstream documentation](https://bun.sh/docs) - the real surface and true flag shapes
