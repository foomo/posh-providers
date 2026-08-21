#### Hazards

**Every task is arbitrary shell, and some of it runs as root.** Each entry in
`cmds`, `precondition` and every transitive `deps` entry is executed with
`sh -c`; a command starting with `sudo ` is re-spawned as `sudo sh -c ...`, so
running a task can escalate to root and may block on a password prompt. The task
name tells you nothing about what it does - **read the task's `cmds` in the
config before running it**, and treat an unfamiliar name as unreviewed code.
There is no dry-run, no confirmation unless the task itself defines one, and no
allow-list.

**A task's dependencies run first, recursively, and there is no cycle guard.**
`deps` are executed before `cmds`, each dependency running its own
preconditions, prompt, deps and commands. So one name can fan out into many
shell commands from other tasks - including hidden ones, which are excluded from
completion but perfectly runnable as dependencies. Two tasks that depend on each
other will recurse until the stack gives out.

**`prompt` is the only interactive brake, and it is opt-in.** When a task sets
it, execution stops on a yes/no confirmation that requires a terminal - an agent
cannot answer it, and declining exits successfully with nothing run. Tasks
without a `prompt` execute immediately. Do not infer safety from the presence of
a confirmation on some other task.

**Commands inherit posh's environment, stdin, stdout and stderr directly**, plus
the task's own `env`. They are not sandboxed and not captured: an interactive
command inside a task will take over the terminal, and a long-running one blocks
until it exits. `dir` changes the working directory for the task's commands.

**Execution stops at the first failing command, leaving earlier ones applied.**
There is no rollback, so a partially-run task can leave the project in a state
neither "before" nor "after".

#### Behaviour

**`precondition` is a skip test, not a guard - and the sense is easy to get
backwards.** Each precondition command is run in order, and the *first one that
succeeds* aborts the whole task as a no-op: the task is considered already
satisfied, and its prompt, deps and cmds are skipped. A precondition that fails
merely moves on to the next. So a precondition that succeeds means "nothing to
do here", and the task reports success without running anything.

Tasks come from two places that are merged: the `tasks` map in the posh config,
and one task per `*.yaml` file in the configured `path` (the file's basename is
the task name). **A file task silently takes precedence over a config task of
the same name.** The merged set is cached for the session, so a task file added
or edited after the shell started is not picked up until the cache is cleared.

`hidden: true` only removes a task from tab completion. It remains runnable by
name and as a dependency - it is a tidiness flag, not access control.

Escalation is decided solely by whether a command string starts with `sudo `,
which is re-spawned as `sudo sh -c ...`. There is no per-task flag for it: the
privilege is visible only in the command text itself.

The confirmation field is `prompt:`. Unknown keys in a task are silently
ignored, so a misspelling means the task runs with no confirmation at all -
the dangerous direction to be wrong in.

#### Configuration

Config key `task` by default, overridable via `WithConfigKey` (and the command
renameable via `CommandWithName`) - note the config-key option here is
`WithConfigKey`, not the `CommandWithConfigKey` some siblings use. Confirm both
against the project's own posh config. Tasks live under `task.tasks`, with
`task.path` pointing at a directory of one-task-per-file YAML.

Field shapes: [`arbitrary/task/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/arbitrary/task/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

The config is the whole surface: there is no way to pass a command from the
prompt, only to name a task. Reading the `cmds`, `deps` and `precondition` of
the task you intend to run is the only way to know what will execute.

#### Examples

```bash
# Runs the task's preconditions, prompt, deps, then its cmds - read them first
posh execute task init
```

#### References

- [Provider README](https://github.com/foomo/posh-providers/blob/main/arbitrary/task/README.md) - plugin wiring; its config sample uses stale key names
