#### Hazards

Both verbs manage **background processes that outlive the command.** `connect`
returns as soon as the forward is up; the `kubectl port-forward` process keeps
running in the posh session, holding a local port. Nothing here is destructive to
the cluster - a port forward only reads - but an agent that calls `connect` and
walks away has left a process and a bound port behind, and a second `connect` for
the same name logs `already running` rather than starting a duplicate.

Each named forward is registered as a gokazi task called `kubeforward.<name>`, so
this command shares a process registry with the `gokazi` command. `gokazi list`
is how you see whether a forward is actually up, and a bare `gokazi stop` stops
these forwards too - they are not isolated from it. Prefer `disconnect` here,
which resolves names against this provider's own config.

**`disconnect` with no argument stops every configured forward**, not nothing:
the name argument is optional and omitting it falls back to the full list from the
config. That is a wider blast radius than the usage block suggests, so pass the
name explicitly unless stopping everything is the intent.

`connect --debug` changes the control flow rather than only the verbosity. In
debug mode the process runs in the **foreground** attached to the terminal, so the
call blocks until the forward dies - and because it returns right after that one
process, a `connect --debug a b` only ever starts `a`. An agent should not use
`--debug`: it will hang waiting on a process that never exits on its own.

After starting a forward, `connect` sleeps one second and then verifies the task
is running, so a forward that fails fast (port already bound, target missing)
surfaces as an error here; one that dies later does not, and only `gokazi list`
will show it.

The names are not free text - they are the keys of this provider's own config
key, and an unknown name is rejected with `port forward <name> not found` before
kubectl is invoked. The cluster each name targets comes from that same config,
and its kubeconfig is resolved through the kubectl provider, so the cluster's
kubeconfig must already exist (fetched by `gcloud`, `az`, `teleport`, ...) or
kubectl will fail to connect. `--profile` selects which kubeconfig subdirectory is
used and does not reach kubectl as a flag.

#### Configuration

Config key `kubeforward` by default, renameable via `CommandWithConfigKey` (and the
command itself via `CommandWithName`), so confirm both against the project's own
posh config. The key is a map of forward name to target, and those keys *are* the
accepted arguments - read it to learn the valid names, the cluster each one hits,
and the local port it binds.

Field shapes: [`kubernetes/kubeforward/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubeforward/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same
schema is bundled under in the project's `posh.schema.json`.

Two things the schema cannot say: `port` is a `"<local>:<remote>"` mapping, so it
is the left-hand side that gets bound on the machine running posh, and
`description` only changes the label shown in `gokazi list` and the prompt check.

#### Examples

```bash
# Start one named forward; the name comes from the kubeforward config key
posh execute kubeforward connect my-database

# Several at once
posh execute kubeforward connect my-database my-cache

# Stop one - naming it, since a bare disconnect stops all of them
posh execute kubeforward disconnect my-database

# Check what is actually running
posh execute gokazi list
```

#### References

- [kubectl port-forward](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#port-forward)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/kubernetes/kubeforward/README.md)
