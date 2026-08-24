#### Hazards

**`disconnect` with no name closes every tunnel of that kind.** The name is
optional, and omitting it substitutes the full list from the config, so
`{{cmd}} cluster disconnect` tears down every cluster tunnel. Both loops return
on the first failure, leaving the tunnels before it already closed.

**Disconnecting kills a process, and it matches on hostname alone.** It kills
the first process named `cloudflared` whose command line contains
`--hostname <hostname>`; the port is never matched on, so two entries sharing a
hostname are indistinguishable. It also reaches tunnels this project did not
start. There is no confirmation.

**`cluster kubeconfig` overwrites kubectl's file for that cluster without
asking.** It fetches a document from 1Password and writes it there (mode
`0600`), replacing a config written by `gcloud`, `az`, or by hand. It needs a
working 1Password session; that is the only verb here that reads a secret.

`status` and `connect` are otherwise non-destructive. `connect` refuses a second
tunnel for a hostname that already has one.

#### Behaviour

`status` is not scoped to this project. It lists **every** running `cloudflared`
process on the machine by pid and command line, and does not resolve them back
to config names. Read the `--hostname` in each to tell which is which.

The tunnels are **not** managed through gokazi. `New` registers a task per
cluster and per database, but no verb here starts or stops them - every one goes
to `cloudflared` directly. Those entries are inert: `gokazi` will list them and
they will never show as running.

Connections outlive the shell: `cloudflared` is started with `Setpgid`, so a
tunnel survives the shell exiting.

Cluster and database names are separate namespaces.

#### Configuration

The `beam` key of this project's posh config holds the `clusters` and
`databases` maps, whose keys are the names the command accepts.

Field shapes: [`foomo/beam/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/beam/config.schema.json).

What the schema cannot say: `port` is the local side only - what `cloudflared`
binds on `127.0.0.1`, so two entries must not share one - while `hostname` is
the remote target and is what `disconnect` matches. Every `$PORT` in the
1Password `kubeconfig` document is replaced with that cluster's `port` before
the file is written.

#### Examples

```bash
posh execute {{cmd}} cluster connect my-cluster
posh execute {{cmd}} cluster kubeconfig my-cluster
posh execute {{cmd}} cluster disconnect my-cluster
```

#### References

- https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/use-cases/ssh/
- `foomo/beam/README.md` for plugin wiring; its snippet does not compile as written
