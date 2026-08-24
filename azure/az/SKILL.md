#### Hazards

`raw` is the whole Azure CLI, not a leaf: it drops the word `raw` and hands the
rest to the real `az` binary, so `{{cmd}} raw group delete --name x` reaches
upstream with every side effect it has and nothing from posh in front of it. Only
run mutating verbs through it with approval.

`vault` is the destructive corner. Its `delete` verbs remove live key vault
material and its `set`/`create` verbs overwrite or add one, none of them
prompting. `show` prints the secret value to stdout - read-only, but not safe to
log. Only `list` and `show` are safe unattended, and a local target and a
production one differ by one argument.

`login` needs a browser unless given `--service-principal <name>`, which logs in
non-interactively from config credentials - the only variant an agent can
complete, and it puts a plaintext secret on the `az` command line. Otherwise ask
a human. `logout` drops those credentials for the rest of the session.

`kubeconfig` runs `aks get-credentials --overwrite-existing`, silently replacing
any existing entry, then shells out to `kubelogin convert-kubeconfig`, which must
be installed separately - unchecked, so it fails after the credentials are
written. Run `login` first. It is the prerequisite for every kubectl-based
command against that cluster.

`artifactory` runs `acr login` on the upstream binary, writing a registry
credential into the local Docker config - docker state, not just Azure state.

#### Behaviour

`AZURE_CONFIG_DIR` points at the configured path, set once at shell start, so
accounts and tokens live per project: one logged in outside posh is invisible
here, and vice versa.

Placeholders resolve against the config, each scoped to the one before it, and
are posh-side labels - the `name` and `resourceGroup` beneath them reach Azure.

`--profile` never reaches `az`; it picks the subdirectory of the kubectl
provider's config path the file lands in, so it must match what the consuming
command reads.

`login` and the `vault` verbs append arguments after `--` to the underlying call.
`artifactory` and `kubeconfig` ignore them.

#### Configuration

The `az` key holds the tenant, the service principals and the subscription tree
the placeholders resolve against. Its `proxy` field is read by nothing here.

Field shapes: [`azure/az/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/azure/az/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} login --service-principal <name>
posh execute {{cmd}} kubeconfig <subscription> <cluster> --profile azure
posh execute {{cmd}} vault <subscription> <vault> secret list
```

#### References

- https://learn.microsoft.com/en-us/cli/azure/reference-index
- https://azure.github.io/kubelogin/ - the `convert-kubeconfig` step `kubeconfig` needs
