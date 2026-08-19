#### Hazards

**`disconnect` with no name closes every tunnel of that kind.** The name is
optional, and when it is omitted the provider substitutes the full list from the
config - so `beam cluster disconnect` tears down every cluster tunnel and
`beam database disconnect` every database tunnel. The plural case is the
default, not an error. Both iterate and return on the first failure, leaving the
tunnels before it already closed.

**Disconnecting kills a process, and it matches on hostname alone.**
`cloudflared.Disconnect` scans every process named `cloudflared` and kills the
first whose command line contains `--hostname <hostname>` - the configured port
is logged but never matched on. Two entries sharing a hostname on different
ports are therefore indistinguishable: disconnecting either kills whichever
process is found first. It also reaches tunnels this project did not start,
including one opened by hand or by the `cloudflared` provider. There is no
confirmation step.

**`cluster kubeconfig` overwrites kubectl's file for that cluster without
asking.** It fetches a document from 1Password and writes it straight to
`<kubectl configPath>/<cluster>.yaml` (mode `0600`), replacing whatever was
there - a config written by `gcloud`, `az`, or by hand. It needs a working
1Password session; that is the only step here that reads a secret. Run it before
any kubectl-based command against that cluster, and know that it is the file
those commands will then use.

**An unknown name is not rejected - it silently does nothing useful.** There is
no `Validate`, and `GetCluster`/`GetDatabase` return a zero-value struct for a
name that is not in the config rather than an error. So a typo reaches
`connect` as an empty hostname and port `0`: the tunnel is opened against
nothing, and `disconnect` matches `--hostname ` against every cloudflared
process. Check the name against the config before running either verb; tab
completion is the reliable source.

`connect` refuses to open a second tunnel for a hostname that already has one,
returning `connection already exists` as an error rather than succeeding
quietly. That check is by hostname, so it is subject to the same collision as
`disconnect`.

#### Behaviour

`status` is not scoped to this project. It lists **every** running `cloudflared`
process on the machine by pid and command line, including tunnels opened by the
`cloudflared` provider or by hand, and it does not resolve them back to beam
config names. Read the `--hostname` in each command line to tell which is which.

The tunnels are **not** managed through gokazi, unlike the other
process-starting providers here. `beam.New` does register a task per cluster and
per database (`beam.cluster.<name>`, `beam.database.<name>`), but nothing in this
command ever starts or stops them - every verb goes to `cloudflared` directly.
Those registry entries are inert: the `gokazi` command will list them, and they
will never show as running. Do not use `gokazi` to manage a beam tunnel.

Connections outlive the shell. `cloudflared.Connect` starts the process with
`Setpgid`, detaching it from the posh process group, so a tunnel survives the
shell exiting and must be closed with `disconnect` (or killed by pid).

The names are the keys of the `clusters` and `databases` maps in the project's
posh config - they differ per project and cannot be listed here. Cluster and
database names are separate namespaces; the same name may appear in both and
mean two different tunnels.

#### Configuration

Config key `beam` by default, overridable via `WithConfigKey` (and the command
renameable via `CommandWithName`), so confirm both against the project's own
posh config.

Field shapes: [`foomo/beam/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/beam/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

Two things the schema cannot say. A cluster's `kubeconfig` is a 1Password
document reference, and the document's contents are templated: every `$PORT` in
it is replaced with that cluster's configured `port` before the file is written,
which is how the kubeconfig points at the local end of the tunnel. And `port` is
the local side only - it is what `cloudflared` binds on `127.0.0.1`, so two
entries must not share one, while `hostname` is the remote Cloudflare Access
target and is what identifies the process for `disconnect`.

The command also needs the `kubectl` and `cloudflared` providers wired in
(`NewCommand(l, beam, kubectl, cloudflared)`); `cluster kubeconfig` additionally
needs 1Password, which is supplied to `beam.New` rather than to the command.

#### Examples

```bash
# Open the tunnel, then fetch the kubeconfig that points through it
posh execute beam cluster connect my-cluster
posh execute beam cluster kubeconfig my-cluster

# Every running cloudflared process, not only beam's
posh execute beam status

# Close one; with no name this closes ALL cluster tunnels
posh execute beam cluster disconnect my-cluster
```

#### References

- [Cloudflare Access `cloudflared access tcp`](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/use-cases/ssh/) - what each tunnel runs
- [Provider README](https://github.com/foomo/posh-providers/blob/main/foomo/beam/README.md) - plugin wiring and a sample config
