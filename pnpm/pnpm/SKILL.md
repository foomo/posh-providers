#### Hazards

**`run <script>` executes whatever that script is defined as in `package.json`.**
Its blast radius is that file's contents, not anything visible on the command
line - read the script before running it. `add` and `install` rewrite
`package.json`/the lockfile and fetch packages from the network, running
dependency lifecycle hooks. `audit` and `list` are read-only.

**The tree is not the surface.** Every node falls back to the real binary,
including the root, so subcommands this tree does not list - `remove`, `update`,
`publish`, `dlx`, `exec` - are still reachable by typing them and behave exactly
as upstream does. Treat this as "all of the package manager", not the six listed
verbs.

**`workspace <path> <verb>` runs the verb with the working directory set to that
package**, so the same `add` or `install` affects a different `package.json`
depending on one argument. The top-level forms run in the project root instead.
Check which of the two you are using before any mutating verb.

#### Behaviour

**Workspace path completion depends on `pnpm-workspace.yaml`.** With the file
present, each `packages` entry (trailing `/*` stripped) is walked for nested
`package.json` files. Without it, discovery scans the project root, so a
single-package project offers `.` and nothing else.

`--recursive` is declared on the root only, but every leaf forwards parsed
flags, so it reaches the binary wherever it is typed.

Script names come from the target directory's `package.json`; workspace paths
from `pnpm-workspace.yaml`. Both lists are cached per session, so a newly added
package or script needs a `cache clear` before it is offered.

Nothing after a `--` separator is forwarded - neither additional args nor
additional flags - so that escape hatch does not work here. Put everything
before it.

#### Configuration

No config key, no schema - nothing to set in the project's posh config. What is
reachable is decided by the project's `pnpm-workspace.yaml` and its
`package.json` files.

The command cannot be renamed and takes no functional options: `NewCommand`
accepts none, and the name is hardcoded rather than read from an option.

#### References

- [Upstream CLI reference](https://pnpm.io/cli/add) - the real surface behind the passthrough
- [Provider README](https://github.com/foomo/posh-providers/blob/main/pnpm/pnpm/README.md)
