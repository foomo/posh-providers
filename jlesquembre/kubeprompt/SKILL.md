#### Hazards

**This opens an interactive shell and does not return.** `kube-prompt` is a
full-screen interactive kubectl prompt; running it blocks until a human exits
it, and it needs a real terminal. It is not usable unattended - an agent that
runs it will hang rather than get output. There is no non-interactive mode
reachable from here.

**Everything typed inside that prompt runs against the selected cluster**, with
the whole kubectl surface available - `delete`, `apply`, `drain` and the rest.
The cluster argument is the only thing scoping the session, and once the prompt
is open nothing further is visible to posh: no logging, no confirmation, no
record of what was run. Confirm the cluster before opening a session against
anything but a local one.

**An unknown cluster name is not rejected.** There is no `Validate`, and
`kubectl.Cluster(name)` constructs a value rather than looking one up, so any
string is accepted and turned into a kubeconfig path. A typo therefore starts
`kube-prompt` pointed at a file that does not exist, and the failure surfaces
from the tool rather than from posh. The completion list is the reliable source
of valid names.

#### Behaviour

The command is `kubeprompt` but the binary it runs is **`kube-prompt`** - it
must be on `PATH` under that name, and this provider offers no option to change
the invocation.

This provider does not create kubeconfigs; it selects among the ones
`kubernetes/kubectl` manages. A cluster is selectable because some other
provider - `gcloud`, `az`, `beam`, `k3d` - wrote `<configPath>/<cluster>.yaml`.
If the cluster you want is missing, run that provider first.

`--profile` does not reach `kube-prompt`. It selects which subdirectory of
kubectl's config path the kubeconfig is read from
(`<configPath>/<profile>/<cluster>.yaml` rather than
`<configPath>/<cluster>.yaml`), and its suggestions only appear once a cluster
argument is present. The kubeconfig is passed as a `KUBECONFIG` environment
variable scoped to this one invocation, so it does not disturb the ambient value
other commands in the shell see.

Only arguments after a `--` separator are forwarded to `kube-prompt`; the tree
itself declares no flags beyond `--profile`, which the provider consumes.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type, no
`config.schema.json` and no `config.base.json`, so there is nothing to set in
the project's posh config and no entry for it in `posh.schema.json`. The
relevant configuration belongs to the `kubernetes/kubectl` provider it is
constructed with, whose `configPath` decides which clusters are selectable: see
[`kubernetes/kubectl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubectl/config.schema.json).

The command cannot be renamed. A `CommandOption` type is declared and accepted
by `NewCommand`, but no options exist, so it is always `kubeprompt`.

#### Examples

```bash
# Opens an interactive session - blocks until exited, needs a terminal
posh execute kubeprompt my-cluster

# Same cluster, kubeconfig read from a profile subdirectory
posh execute kubeprompt my-cluster --profile azure
```

#### References

- [kube-prompt](https://github.com/c-bata/kube-prompt) - the interactive prompt this wraps
- [`kubernetes/kubectl`](https://github.com/foomo/posh-providers/blob/main/kubernetes/kubectl/README.md) - owns the kubeconfigs this command selects from
