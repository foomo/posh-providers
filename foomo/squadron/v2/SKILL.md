#### Hazards

The three leading placeholders are resolved against each other, left to right:
the fleets are those configured for the chosen cluster, and the units are
resolved against the cluster and fleet already picked. The squadron names come
off disk rather than from the cluster - see Configuration below for where each
of the three actually comes from. `all` is accepted in place of a single squadron
name to target every one of them. Run
`squadron <cluster> <fleet> <squadron> list` to see the units within a squadron;
naming no unit targets all of them.

Read-only, safe to run for orientation: `status`, `diff`, `template`, `config`,
`list`, `schema`.

Mutating: `up`, `down` and `rollback` change a live cluster. `build` and `push`
build and publish container images - they do not touch the cluster, but `push`
writes to a shared registry and `up --push` does both.

Where this project marks a cluster as requiring confirmation, `up`, `down` and
`rollback` are refused outright under an agent rather than prompting: the prompt
needs a terminal, and the project singled out that cluster as needing a human.
No flag overrides this - ask the user to run it in an interactive shell.

Where a cluster is marked for notification, `up`, `down` and `rollback` post to
Slack, tagged with the current git ref and git user name. The `--slack` flag
forces this on for clusters that do not have it set.

`--tag` sets the image tag via the `TAG` environment variable rather than a
squadron flag; `--tags` is unrelated and filters units by their config tags.
Flags after a `--` separator are passed through to helm.

#### Configuration

Two separate layers, and they answer different questions:

The **posh provider config** is the `squadron` key of this project's posh config.
Read that key for the cluster and fleet names, and its schema - referenced from
the config's own `$schema`, under
`https://github.com/foomo/posh-providers/foomo/squadron` - for the field shape;
both are authoritative in a way this document cannot be. Two behaviours the
schema cannot express: a cluster's `confirm` is what makes the mutating verbs
refuse to run under an agent, and the selectable clusters are the intersection
with the configured kubectl clusters, so a cluster listed here without a
matching kubectl entry is not offered.

**squadron's own config** lives in `squadron.yaml` files under `path` and is
read by the `squadron` binary, not by posh. The offered squadron names are just
the non-hidden directories under `path`, and the units come from the
`squadron.yaml` files within them. None of this appears in the posh config, so to
find a squadron or unit name look on disk or run `list` - do not expect the
`squadron` key to list them.

That second layer is assembled per invocation from a cascade of increasingly
specific filenames - `squadron.yaml`, then `squadron.<fleet>.yaml`,
`squadron.<cluster>.yaml` and `squadron.<cluster>.<fleet>.yaml`, at root,
squadron and unit level - each with an optional `.override.yaml` counterpart.
`--no-override` drops every override file from that set, which is the fastest
way to tell whether a local override is what is affecting a command's output.

#### Examples

```bash
posh execute squadron prod default all status
posh execute squadron prod default frontend diff
posh execute squadron dev default frontend up --build --tag=my-branch
```

#### References

- https://github.com/foomo/squadron for the `squadron.yaml` format
- `foomo/squadron/v2/README.md` for plugin wiring and the posh config sample
