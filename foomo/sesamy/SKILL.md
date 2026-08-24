#### Hazards

**Omitting the `config` argument runs the verb against *every* configured set.**
It is optional on every leaf, and when absent the provider loops over the full
list of config keys. So `provision all` with no argument provisions every Google
Tag Manager container the project knows about - likely including production -
rather than erroring. Naming a set is what narrows it. The loop stops at the
first failure, leaving earlier sets already provisioned.

**`provision` writes to live Google Tag Manager containers**, pushing the
config's tags, triggers and variables; `all` does web then server in sequence.
There is no confirmation and no dry-run flag on the posh side. `diff` shows what
would change and `list` enumerates existing resources - prefer both first, and
treat `provision` as needing explicit approval.

**Every subcommand resolves 1Password secrets into the config before running**,
so even read-only verbs can trigger a prompt and need a signed-in session. The
rendered config is piped on **stdin** (`--config -`), keeping resolved secrets
out of the process list and off disk.

**`config` prints the merged and rendered configuration**, so its output holds
whatever the references resolved to. Not something to paste into a log.

`open` launches a browser console, so it needs a graphical session.

#### Behaviour

A config set is a **list of YAML files merged in order**, later overriding
earlier. The config a command runs with therefore exists nowhere on disk as a
single file - reading one listed file tells you only part of it. Run
`{{cmd}} config <name>` for the effective result.

An unknown name is rejected with `invalid config key: <name>` before the
upstream binary runs.

On `open` and on `list`, the target or resource comes *before* the optional
config name.

`typescript` writes generated definitions wherever the merged config points - a
filesystem side effect driven by config, not by any argument here.

#### Configuration

The `sesamy` key of this project's posh config is a **flat map of set name to a
list of YAML file paths**, not an object with fields. The keys are what the
`config` argument accepts.

Field shapes: [`foomo/sesamy/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/sesamy/config.schema.json).

The `onepassword` provider must be wired in; it renders the secret references.

#### Examples

```bash
posh execute {{cmd}} config my-set
posh execute {{cmd}} diff web my-set
posh execute {{cmd}} provision all my-set
```

#### References

- https://github.com/foomo/sesamy-cli
- `foomo/sesamy/README.md` for plugin wiring and a sample config
