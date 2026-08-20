#### Hazards

This command manages the **shared** background-process registry, not processes it
owns. Other providers register their tasks into the same `*gokazi.Gokazi` instance
- `kubeforward` port forwards, plus `ssh`, `dockprox` and `gost` tunnels - so
everything they started is visible and stoppable here. A task name like
`kubeforward.my-database` belongs to another command; prefer that command's own
`disconnect`, which resolves names against its config and logs what it is doing.

`list` is read-only and the safe way to see what is running: it prints PID, name
and whether each task is actually alive.

**`stop` with no name stops every running task in the registry**, across every
provider - the name argument is optional and omitting it widens the scope rather
than narrowing it, which the usage block cannot convey. Naming one or more tasks
stops exactly those. When you mean one process, prefer the owning provider's own
stop verb, which resolves the name against its config and logs what it does.

Stopping is a signal to a live process, which makes it disruptive rather than
destructive - nothing on disk or in a cluster is touched, but whatever depended on
that tunnel or forward loses it, and a task that another command is mid-way
through using will fail. It needs a human decision when the session is not yours
to disturb. `stop` skips tasks already dead and logs per-task failures without
aborting, so a partial result is normal and the exit status does not tell you
which tasks stopped - read the log lines or re-run `list`.

If the project registers `TasksChecker`, the prompt lists every task each tick
with a filled marker for running and a hollow one for stopped, so a task
disappearing from `list` is visible there too.

Process cleanup on shell exit is opt-in through the config, and it applies to the
whole registry: with it enabled, closing the last posh session interrupts every
task any provider started. With it disabled, background processes survive the
shell and an agent that started one has leaked it.

#### Configuration

Config key `gokazi` by default, renameable via `CommandWithConfigKey` (and the
command itself via `CommandWithName`), so confirm both against the project's own
posh config.

Field shapes: [`foomo/gokazi/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/gokazi/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same
schema is bundled under in the project's `posh.schema.json`.

The single `cleanup` field is what decides whether tasks outlive the shell, so it
is worth reading before assuming a forward will still be up in the next session.
Nothing in this config declares tasks - the task list is contributed at runtime by
whichever providers the project wires up, which is why `list` is the only reliable
inventory.

#### Examples

```bash
# The safe inventory - what is running, and its PID
posh execute gokazi list

# Stop everything in the registry, across every provider
posh execute gokazi stop

# Stop one named task
posh execute gokazi stop kubeforward.my-database

# Or use the owning provider, which validates the name against its config
posh execute kubeforward disconnect my-database
```

#### References

- [gokazi](https://github.com/foomo/gokazi)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/foomo/gokazi/README.md)
