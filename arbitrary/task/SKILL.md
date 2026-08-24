#### Hazards

**Every entry here is arbitrary shell, and some of it runs as root.** Every
string in `cmds`, `precondition` and each transitive `deps` entry runs under
`sh -c`, and one starting with `sudo ` is re-spawned as `sudo sh -c ...`, so
running one can escalate to root and block on a password prompt. The name tells
you nothing - **read its `cmds` first** and treat an unfamiliar name as
unreviewed code. No dry-run, no allow-list, no confirmation unless defined.

**Dependencies run first, recursively, with no cycle guard.** Each runs its own
preconditions, prompt, deps and commands, so one name fans out into commands
belonging to others - including hidden ones. Two entries depending on each other
recurse until the stack gives out.

**`prompt` is the only interactive brake, and it is opt-in.** When set, execution
stops on a yes/no confirmation requiring a terminal - an agent cannot answer it,
and declining exits successfully with nothing run. Entries without one execute
immediately. Unknown keys are silently ignored, so misspelling `prompt:` means no
confirmation at all.

**Commands inherit posh's environment, stdin, stdout and stderr directly**, plus
the entry's `env`, and are not sandboxed - an interactive one takes over the
terminal, a long-running one blocks. Execution stops at the first failure with no
rollback, leaving earlier commands applied.

#### Behaviour

**`precondition` is a skip test, not a guard, and the sense inverts easily.**
Each runs in order and the *first one that succeeds* aborts everything as a
no-op - prompt, deps and cmds skipped, reported as success. One that fails merely
moves on to the next.

The runnable set merges the `tasks` map with one `*.yaml` file per entry in the
configured `path`, named for the file's basename. **A file silently overrides a
configured entry of the same name**; only `.yaml` is loaded, not `.yml`. The
merged set is cached for the session, so later edits are not picked up.

`hidden: true` only removes an entry from completion - it stays runnable by name
and as a dependency. Tidiness, not access control.

#### Configuration

Entries live under the `task` key's `tasks` map, with `path` pointing at a
directory of one-per-file YAML. Only a name can be passed from the prompt, never
a command.

Field shapes: [`arbitrary/task/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/arbitrary/task/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} init
```

#### References

- [Provider README](https://github.com/foomo/posh-providers/blob/main/arbitrary/task/README.md) - its config sample uses stale key names
