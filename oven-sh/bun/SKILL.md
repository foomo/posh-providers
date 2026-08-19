#### Hazards

**Several verbs modify the project, and `run` executes arbitrary package
scripts.** `install`, `add`, `update` and `remove` rewrite `package.json` and
the lockfile and fetch packages from the network; `patch` prepares a package for
patching; `x` **downloads and executes a package binary, installing it if it is
not present**. `run <script>` executes whatever that script is defined as in
`package.json`, so its blast radius is that file's contents, not anything
visible here. Read the script before running it. Only `outdated` and `pm` (with
a read-only subcommand) are safe unattended.

**The tree is not the surface.** The root and every leaf forward to the real
`bun` binary, so subcommands this tree does not list - `build`, `init`, `link`,
`repl` and the rest - are still reachable by typing them, and behave as upstream
does. Treat this as "all of bun", not as the listed verbs.

**`bun workspace <path> <verb>` runs the verb with the working directory set to
that package**, so the same `install` or `remove` affects a different
`package.json` depending on one argument. The top-level `bun install` runs in
the project root instead. Check which of the two forms you are using before any
mutating verb.

**`--watch` and `--hot` do not return.** Both keep the process running and
restart it on file change, so a command carrying either blocks until
interrupted - not usable unattended. `--inspect-wait` and `--inspect-brk` also
block, waiting for a debugger to attach.

#### Behaviour

Flags are passed through **verbatim as typed**, not re-rendered from the parsed
flag set, so the declarations in this tree only drive completion and are not
what reaches bun. That matters because several value-taking bun options are
declared here as booleans - `--install`, `--preload`, `--omit`, `--print` - so
completion offers them as bare switches, while `--omit=dev` typed in full is
forwarded correctly. Trust bun's own documentation for a flag's shape, not this
tree's suggestion.

The flag list is also partial and hand-maintained: many upstream options are
present but commented out in the source (`--port`, `--config`, `--cwd`,
`--silent`, `--eval`, ...). They are absent from completion yet still work when
typed, since forwarding is verbatim.

**Nothing after a `--` separator is forwarded.** Neither additional args nor
additional flags are passed on - unusual among these providers, and it means the
`--` escape hatch does not work here. Put everything before it.

Workspace paths are discovered from the root `package.json`: each entry in its
`workspaces.packages` (with a trailing `/*` stripped) is walked for nested
`package.json` files, and their directories become the completions. With no
workspaces declared, the walk starts at `.` instead. Script names for `run`
come from the target directory's `package.json`. Both lists are cached per
session, so a newly added package or script needs a `cache clear` before it is
offered.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type, no
`config.schema.json` and no `config.base.json`, so there is nothing to set in
the project's posh config and no entry for it in `posh.schema.json`. What is
reachable is decided by the project's `package.json` files; how bun behaves is
decided by its own `bunfig.toml`.

The command cannot be renamed - it is always `bun`. The only option is
`CommandWithExecBun`, which swaps the binary invocation.

#### Examples

```bash
# Runs the script as defined in package.json - read it first
posh execute bun run build

# Same verb, but scoped to one workspace package
posh execute bun workspace packages/api install

# Downloads and executes a package binary if not already present
posh execute bun x some-cli
```

#### References

- [Bun documentation](https://bun.sh/docs) - the real command surface and the true shape of each flag
- [Provider README](https://github.com/foomo/posh-providers/blob/main/oven-sh/bun/README.md)
