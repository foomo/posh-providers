#### Hazards

**`run <script>` executes whatever that script is defined as in `package.json`.**
Its blast radius is that file's contents, not anything visible on the command
line - read the script before running it. `add` and `install` rewrite
`package.json`/the lockfile and fetch packages from the network. `audit` and
`list` are read-only.

**The tree is not the surface.** Every node falls back to the real `pnpm`
binary, including the root, so subcommands this tree does not list - `remove`,
`update`, `publish`, `dlx`, `exec` - are still reachable by typing them and
behave exactly as upstream does. Treat this as "all of pnpm", not as the six
listed verbs.

**`pnpm workspace <path> <verb>` runs the verb with the working directory set to
that package**, so the same `add` or `install` affects a different `package.json`
depending on one argument. The top-level forms run in the project root instead.
Check which of the two you are using before any mutating verb.

#### Behaviour

**Workspace path completion is broken when there is no `pnpm-workspace.yaml`.**
The lookup tests `errors.Is(err, os.ErrExist)` where it means `ErrNotExist`, so
the "file absent" case falls into the error branch and returns an empty list
instead of falling back to scanning `.`. In a single-package project the
`workspace` subtree therefore offers no paths at all. When the file *is*
present, discovery works: each `packages` entry (with a trailing `/*` stripped)
is walked for nested `package.json` files.

**`run` drops your flags.** The two `run` leaves call a helper that execs
`pnpm run <script>` with only the directory set - unlike every other node, which
forwards `r.Flags()`. So `pnpm run build --silent` silently loses `--silent`.
Flags reach pnpm on the other verbs and through the root passthrough, just not
here.

`--recursive` is declared on the root only. It is forwarded by the fallback
paths, so `pnpm install --recursive` works, but it has no effect on the two
`run` leaves for the reason above.

Script names come from the target directory's `package.json`; workspace paths
come from `pnpm-workspace.yaml`. Both lists are cached per session, so a newly
added package or script needs a `cache clear` before it is offered.

Nothing after a `--` separator is forwarded - neither additional args nor
additional flags - so that escape hatch does not work here. Put everything
before it.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type, no
`config.schema.json` and no `config.base.json`, so there is nothing to set in
the project's posh config and no entry for it in `posh.schema.json`. What is
reachable is decided by the project's `pnpm-workspace.yaml` and its
`package.json` files.

The command cannot be renamed - it is always `pnpm` - and takes no functional
options. A `CommandOption` type is declared but `NewCommand` does not accept
any.

#### Examples

```bash
# Runs the script as defined in package.json - read it first
posh execute pnpm run build

# Same verb, scoped to one workspace package
posh execute pnpm workspace packages/api add lodash

# Unlisted subcommand, reaching pnpm directly
posh execute pnpm remove lodash
```

#### References

- [pnpm CLI](https://pnpm.io/cli/add) - the real command surface behind the passthrough
- [Provider README](https://github.com/foomo/posh-providers/blob/main/pnpm/pnpm/README.md) - plugin wiring only
