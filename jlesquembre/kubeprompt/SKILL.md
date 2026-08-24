#### Hazards

**This opens an interactive shell and does not return.** `kube-prompt` is a
full-screen interactive kubectl prompt; running it blocks until a human exits it
and it needs a real terminal, so an agent that runs it hangs rather than getting
output. There is no non-interactive mode here; use `kubectl` instead.

**Everything typed inside that prompt runs against the selected cluster**, with
the whole kubectl surface available - `delete`, `apply`, `drain` and the rest.
The cluster argument is the only thing scoping the session, and once the prompt
is open nothing is visible to posh: no logging, no confirmation, no record of what
was run. Confirm the cluster before opening one on anything but a local cluster.

**An unknown cluster name is not rejected.** There is no `Validate`, and
`kubectl.Cluster(name)` constructs a value rather than looking one up, so any
string is accepted and turned into a kubeconfig path. A typo starts the binary
pointed at a file that does not exist, and the failure surfaces from the tool
rather than from posh. The completion list is the reliable source of valid names.

#### Behaviour

The command is `{{cmd}}` but the binary it runs is **`kube-prompt`** - it must be
on `PATH` under that name, and there is no option to change the invocation. The
command cannot be renamed either: a `CommandOption` type is declared and accepted
by `NewCommand`, but no options exist.

This provider does not create kubeconfigs; it selects among the ones
`kubernetes/kubectl` manages. A cluster is selectable because some other provider
- `gcloud`, `az`, `beam`, `k3d` - wrote `<configPath>/<cluster>.yaml`. If the one
you want is missing, run that provider first.

`--profile` does not reach the binary. It selects which subdirectory of kubectl's
config path the kubeconfig is read from, and its suggestions only appear once a
cluster argument is present. The kubeconfig is a `KUBECONFIG` variable scoped to
this one invocation.

Only arguments after `--` are forwarded; flags typed on the command are not.

#### Configuration

**No config key of its own** - no `Config` type, no schema. The relevant
configuration belongs to the `kubernetes/kubectl` provider it is constructed with,
whose `configPath` decides which clusters are selectable:
[`kubernetes/kubectl/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/kubernetes/kubectl/config.schema.json).

#### Examples

```bash
# Blocks until exited, needs a terminal
posh execute {{cmd}} my-cluster --profile azure
```

#### References

- [kube-prompt](https://github.com/c-bata/kube-prompt) - the interactive prompt this wraps
- [`kubernetes/kubectl`](https://github.com/foomo/posh-providers/blob/main/kubernetes/kubectl/README.md) - owns the kubeconfigs this command selects from
