#### Hazards

Unrecognised arguments fall through to the real `gcloud` binary, so this command
is a full passthrough and not limited to the three subcommands listed above -
`gcloud projects list` or `gcloud compute instances list` work and reach the
upstream CLI unchanged. The named subcommands only exist because they wrap
something posh resolves for you. Any read-only gcloud verb is safe for
orientation; anything you invoke through the passthrough carries exactly the side
effects the upstream CLI gives it, and posh adds no confirmation step.

Every invocation runs with `CLOUDSDK_CONFIG` pointed at this provider's
configured path, so gcloud's own state - the active account, generated
credentials - is kept per project rather than in the user's real home directory.
This is set once when the shell starts, so an account activated here does not
leak into the user's global gcloud config, and an account they logged into
outside posh is not visible in here.

`login` mutates that per-project state by switching the active account, and its
behaviour depends on what it finds on disk. It first looks for a service account
key at `service_account_keys/<account>.json` under the configured path; failing
that, and only if the account has a `key` and this project wired up 1Password, it
downloads the key to that location - writing a credential to the checkout - and
activates it. With neither, it falls back to `gcloud auth login`, which opens a
browser and cannot complete without a graphical session. So for an agent: `login`
against a service-account-backed account is fine, but against an account with no
key and no cached key file it will hang on a browser - ask a human to run it.
The account names come from the `gcloud` config key, not from a live lookup.

`kubeconfig` writes a kubeconfig file and is the prerequisite for every
kubectl-based command against that cluster; run it before anything expects the
cluster to be reachable. The cluster names are the keys of the `gcloud` config
key's cluster map - the value written to the config is the cluster's configured
`name`, `project` and `region`, so the name you pass is the posh-side label, not
necessarily the GKE cluster name. If the cluster names an account, that account
is passed to gcloud as `--account`, which means `kubeconfig` can fail on
authentication if you have not run `login` for it first.

#### Behaviour

`--profile` does not reach gcloud. It selects which kubeconfig file the
credentials are written into - a subdirectory of the kubectl provider's config
path - so it partitions kubeconfigs rather than changing anything about the
gcloud call. It defaults to empty, writing to the top-level kubeconfig; `gcloud`
is offered as a suggestion but any name is accepted and simply becomes a
directory. Whatever you pass here must match what the consuming kubectl-based
command is told to read, or that command will not find the credentials this one
just wrote.

A cluster needs no separate entry in the kubectl provider's config for
`kubeconfig` to succeed: the destination path is derived from the cluster name
and the kubectl provider's `configPath`, so the file is written either way. What
that means in practice is that a typo in the cluster name is not rejected here -
it writes a kubeconfig under the misspelled name, and the failure surfaces later
in whichever command cannot find it.

Arguments after a `--` separator are appended to the underlying gcloud call by
every subcommand, which is how you reach upstream flags the tree does not model.

#### Configuration

The `gcloud` key of this project's posh config holds the account and cluster
names - read that key for the valid values. `gcloud` is only the default key,
overridable via `WithConfigKey`, so confirm the actual one against the project's
own file.

Field shapes: [`google/gcloud/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/google/gcloud/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same
schema is bundled under in the project's `posh.schema.json`.

Three behaviours the schema cannot express: an account's `key` is what decides
whether `login` can run unattended, `configPath` is created on startup and
becomes `CLOUDSDK_CONFIG` for every invocation, and a cluster's key here is only
a label - the `name`, `project` and `region` beneath it are what reach GCP.

#### Examples

```bash
posh execute gcloud kubeconfig <cluster>
posh execute gcloud login <account>
posh execute gcloud projects list
```

#### References

- https://cloud.google.com/sdk/gcloud/reference
- `google/gcloud/README.md` for plugin wiring and the posh config sample
