#### Hazards

`auth init` prompts on stdin for a context name and an API token, so it cannot
run unattended, and writes that token in plaintext inside the checkout - confirm
the path is git-ignored, and never commit it. `-- --access-token <t>` does not
skip the prompt (posh keeps the `--`, so the flag never binds); use
`DIGITALOCEAN_ACCESS_TOKEN` in the environment instead.

The `registry` verbs change state outside the project: the credential lands in
the host's Docker config, so `logout` breaks anything else on the machine pulling
from that registry. `--never-expire` mints a token that never expires where the
default lasts 30 days - a long-lived credential on a shared machine.

`kubeconfig` merges into the target file and switches its current context to the
cluster just saved, so a later bare `kubectl` talks to whatever was saved last -
and no upstream flag reaches this verb to turn that off (see Behaviour).

The root carries no `Execute`, so the rest of the DigitalOcean CLI is
unreachable: nothing here touches droplets, clusters or registries.

#### Behaviour

The generated kubeconfig holds no static credential - upstream writes an `exec`
block fetching a token on demand, so it is useless where the upstream binary is
absent: in a CI image lacking it, it fails on authentication rather than
permissions. That plugin needs the cluster UUID, hence a cluster's `name` is
normally a UUID.

`kubeconfig` forwards additional args but not flags, and `--` does not rescue
them: posh keeps the separator and upstream ignores the extra positional, so
`kubeconfig prod -- --set-current-context=false` neither applies nor errors. The
other verbs forward flags normally.

`<cluster>` completes from the config keys, not the API, so a cluster present in
the account but absent from the config cannot be named; an unknown alias is
rejected up front.

`DIGITALOCEAN_CONFIG` is exported at shell start, pinning the session to that
file's default context. `DIGITALOCEAN_ACCESS_TOKEN` is never exported - the guard
that would set it cannot be satisfied - so set it yourself. The godo client and
`CommandWithAuthTokenProvider` are dead code.

#### Configuration

The `doctl` key's `clusters` map is the whole addressable surface - read it for
the aliases and the cluster ID each maps to. Where the kubeconfig lands is
governed by the `kubectl` provider.

Field shapes: [`digitalocean/doctl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/digitalocean/doctl/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} kubernetes kubeconfig prod --profile ci
```

#### References

- https://docs.digitalocean.com/reference/doctl/reference/kubernetes/cluster/kubeconfig/save/
- `digitalocean/doctl/README.md` for plugin wiring
