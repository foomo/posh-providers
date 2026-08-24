#### Hazards

`up`, `down` and `rollback` change a live cluster. `push` publishes images to a
shared registry, and `up --push` does both. `status`, `diff`, `template`,
`config`, `list` and `schema` are read-only and safe for orientation.

Where this project marks a cluster `confirm`, `up`, `down` and `rollback` are
refused outright under an agent rather than prompting: the prompt needs a
terminal, and the project singled out that cluster as needing a human. No flag
overrides this - ask the user to run it in an interactive shell.

Where a cluster is marked for notification, those same three verbs post to
Slack, tagged with the current git ref and git user name. `--slack` forces this
on for clusters that do not have it set.

Naming no unit targets every unit; `all` in place of a single name in the third
placeholder targets every one of them.

#### Behaviour

The three leading placeholders resolve against each other, left to right: fleets
come from the chosen cluster's config, units from the cluster and fleet already
picked. Squadron names come off disk, not from the config - they are the
non-hidden directories under `path`, and the units come from the `squadron.yaml`
files within them. Selectable clusters are the intersection with the configured
kubectl clusters, so a cluster configured here without a matching kubectl entry
is never offered.

`--tag` sets the image tag via the `TAG` environment variable rather than a
flag; `--tags` is unrelated and filters units by their config tags. Flags after
a `--` separator are passed through to helm.

#### Configuration

The `squadron` key of this project's posh config holds the cluster and fleet
names, `path`, and the per-cluster `confirm` and `notify` switches.

Field shapes: [`foomo/squadron/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/squadron/config.schema.json).

The tool's own config is separate: `squadron.yaml` files under `path`, read by
the `squadron` binary, not by posh. Each invocation assembles a cascade of
increasingly specific filenames - `squadron.yaml`, then `squadron.<fleet>.yaml`,
`squadron.<cluster>.yaml` and `squadron.<cluster>.<fleet>.yaml`, at root,
directory and unit level - each with an optional `.override.yaml` counterpart.
`--no-override` drops every override file, the fastest way to tell whether a
local override is affecting a command's output.

#### Examples

```bash
posh execute {{cmd}} prod default all status
posh execute {{cmd}} prod default frontend diff
posh execute {{cmd}} dev default frontend up --build --tag=my-branch
```

#### References

- https://github.com/foomo/squadron for the `squadron.yaml` format
- `foomo/squadron/v2/README.md` for plugin wiring and the posh config sample
