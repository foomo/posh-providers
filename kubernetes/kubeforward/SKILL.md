#### Hazards

Both verbs manage **background processes that outlive the command.** `connect`
returns as soon as the forward is up; the `kubectl port-forward` process keeps
running in the posh session, holding a local port. Nothing here is destructive to
the cluster - a port forward only reads - but an agent that calls `connect` and
walks away has left a process and a bound port behind. A second `connect` for the
same name logs `already running` rather than duplicating.

Each forward is a gokazi task named `kubeforward.<name>` (that prefix is fixed
even when the command is renamed), so this shares a process registry with the
`gokazi` command: `gokazi list` shows whether a forward is really up, and a bare
`gokazi stop` stops these forwards too.

**`disconnect` with no argument stops every configured forward**, not nothing:
the name is optional and omitting it falls back to the full list from the config.
That is a wider blast radius than the usage block suggests, so name it explicitly
unless stopping everything is the intent.

`connect --debug` changes the control flow, not just the verbosity: the process
runs in the **foreground** attached to the terminal, so the call blocks until the
forward dies - and because it returns right after that one process,
`connect --debug a b` only ever starts `a`. An agent should not use it.

#### Behaviour

After starting a forward, `connect` sleeps one second and verifies the task is
running, so one that fails fast (port already bound, target missing) surfaces as
an error here; one that dies later does not, and only `gokazi list` shows it.

An unknown name is rejected with `port forward <name> not found` before kubectl
runs. The cluster each name targets comes from the config and its kubeconfig is
resolved through the `kubectl` provider, so that kubeconfig must already exist
(fetched by `gcloud`, `az`, `teleport`, ...) or kubectl fails to connect.
`--profile` selects the kubeconfig subdirectory and is not passed on as a flag.

#### Configuration

The `kubeforward` key is a map of forward name to target, and those keys *are*
the accepted arguments - read it to learn the valid names, the cluster each one
hits, and the local port it binds.

Field shapes: [`kubernetes/kubeforward/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubeforward/config.schema.json).

What the schema cannot say: `port` is a `"<local>:<remote>"` mapping, so its
left-hand side is what gets bound locally.

#### References

- [kubectl port-forward](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#port-forward)
- [Provider README](https://github.com/foomo/posh-providers/blob/main/kubernetes/kubeforward/README.md)
