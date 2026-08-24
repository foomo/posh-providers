#### Hazards

The root is a passthrough: an unrecognised argument goes straight to the real
`gcloud` binary, so the listed subcommands are not the surface. Anything invoked
that way carries exactly the side effects upstream gives it, with nothing from
posh in front of it. Read-only verbs are fine for orientation; treat the rest as
needing approval.

`login` switches the active account, and what it does depends on what it finds on
disk. It looks for a service account key at `service_account_keys/<account>.json`
under the configured path; failing that, and only if the account has a `key` and
this project wired up 1Password, it downloads that key there - writing a
credential into the checkout - and activates it. With neither, it falls back to
`auth login`, which opens a browser and cannot complete headless. So a
service-account-backed account is safe unattended; one with no key will hang -
ask a human.

`kubeconfig` writes the kubeconfig every kubectl-based command against that
cluster needs, so run it first. If the cluster names an account, that account is
passed as `--account`, so it fails on authentication unless `login` ran for it.

`docker` adds a credential helper to the local Docker config - machine state,
not project state.

#### Behaviour

`CLOUDSDK_CONFIG` points at the configured path, set once at shell start, so the
active account and generated credentials live per project: an account activated
here does not leak into the user's global config, and one logged in outside posh
is invisible here.

Account and cluster names are keys of the `gcloud` config maps, not a live
lookup, and an unknown one is rejected before anything runs. A cluster key is
only a label - the `name`, `project` and `region` beneath it reach GCP.

`--profile` never reaches `gcloud`; it selects which subdirectory of the kubectl
provider's config path the credentials land in, and nothing cross-checks that the
result is a path `kubectl` will look for, so it must match what the consuming
command reads. Every subcommand appends arguments after `--` to the call.

#### Configuration

The `gcloud` key holds the account and cluster maps the arguments resolve
against. An account's `key` decides whether `login` can run unattended.

Field shapes: [`google/gcloud/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/google/gcloud/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} login <account>
posh execute {{cmd}} kubeconfig <cluster> --profile ci
posh execute {{cmd}} projects list
```

#### References

- https://cloud.google.com/sdk/gcloud/reference
- `google/gcloud/README.md` for plugin wiring and a config sample
