#### Hazards

This command manages the **shared** background-process registry, not processes
it owns. Other providers register tasks into the same instance - `kubeforward`
port forwards, plus `ssh`, `dockprox` and `gost` tunnels - so everything they
started is visible and stoppable here. A task named `kubeforward.my-database`
belongs to another command.

**`stop` with no name stops every running task in the registry**, across every
provider: the name argument is optional and omitting it widens the scope rather
than narrowing it. Naming one or more tasks stops exactly those. When you mean
one process, prefer the owning provider's own stop verb, which resolves the name
against its config and logs what it does.

Stopping signals a live process, so it is disruptive rather than destructive -
nothing on disk or in a cluster is touched, but whatever depended on that tunnel
or forward loses it, and a task another command is mid-way through using will
fail. It needs a human decision when the session is not yours to disturb.

`stop` skips tasks already dead and logs per-task failures without aborting, so
a partial result is normal and the exit status does not say which tasks stopped
- read the log lines or re-run `list`.

`list` is read-only and the safe way to see what is running: PID, name, and
whether each task is actually alive.

#### Behaviour

Process cleanup on shell exit is opt-in through the config and applies to the
whole registry: enabled, closing the last posh session interrupts every task any
provider started; disabled, background processes survive the shell, so an agent
that started one has leaked it.

If the project registers `TasksChecker`, the prompt lists every task each tick,
filled for running and hollow for stopped.

#### Configuration

The `gokazi` key of this project's posh config holds a single `cleanup` field,
which decides whether tasks outlive the shell - worth reading before assuming a
forward will still be up next session.

Field shapes: [`foomo/gokazi/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/gokazi/config.schema.json).

Nothing in this config declares tasks. The task list is contributed at runtime
by whichever providers the project wires up, which is why `list` is the only
reliable inventory.

#### Examples

```bash
posh execute {{cmd}} list
posh execute {{cmd}} stop
posh execute {{cmd}} stop kubeforward.my-database
posh execute kubeforward disconnect my-database
```

#### References

- https://github.com/foomo/gokazi
- `foomo/gokazi/README.md` for plugin wiring and the sibling providers sharing the registry
