#### Hazards

**Omitting the `config` argument runs the verb against *every* configured set.**
It is optional on every leaf, and when absent the provider substitutes the full
list of config keys and loops. So `sesamy provision all` with no argument
provisions every Google Tag Manager container the project knows about - likely
including production - rather than erroring. Naming a set is what narrows it.
The loop stops at the first failure, leaving earlier sets already provisioned.

**`provision` writes to Google Tag Manager.** `provision web`, `provision
server` and `provision all` push the config's tags, triggers and variables into
live GTM containers; `all` does web then server in sequence. There is no
confirmation and no dry-run flag on the posh side. `diff` is the read-only way
to see what would change, and `list` enumerates existing resources - prefer both
before provisioning, and treat `provision` as needing explicit approval.

**Every subcommand resolves 1Password secrets into the config before running.**
The merged YAML is passed through the onepassword provider's `Render`, so any
secret reference in it is fetched - meaning even read-only verbs like `config`,
`tags` and `list` can trigger a 1Password prompt and require a signed-in
session. The rendered config is piped to sesamy on **stdin** (`--config -`), so
resolved secrets do not appear in the process list or on disk. One exception:
when rendering fails, the *unrendered* config is printed to stderr for
debugging - that content still holds the unresolved references, not the secret
values.

**`config` prints the fully merged and rendered configuration**, so its output
contains whatever secrets the references resolved to. It is a debugging aid, not
something to paste into a log or an issue.

#### Behaviour

A config set is a **list of YAML files merged in order**, later files overriding
earlier ones, keyed by a name in the posh config. So the config the command runs
with exists nowhere on disk as a single file - reading any one of the listed
files tells you only part of it. `sesamy config <name>` is the way to see the
effective result.

An unknown config name is rejected with `invalid config key: <name>` before
sesamy runs. The names come from the posh config and are offered by tab
completion.

`open` takes a target - `ga`, `gtm-web` or `gtm-server` - and launches a browser
console, so it needs a graphical session and is not usable unattended. Note the
argument order: the target comes *before* the optional config name.

Flag forwarding is inconsistent between verbs. `list` forwards only the flags
you actually typed (`fs.Visited()`); `config`, `tags`, `typescript`, `open`,
`provision` and `diff` forward `r.Flags()`. Additional args after a `--`
separator are forwarded everywhere; additional *flags* are forwarded nowhere, so
a flag placed after `--` is silently dropped.

The `--verbose` flag is declared on `tags` and `typescript` only; the other
leaves declare their own flag sets.

`typescript` writes generated definitions to wherever the merged config points -
a filesystem side effect driven entirely by config, not by any argument here.

#### Configuration

Config key `sesamy` by default, overridable via `CommandWithConfigKey` (and the
command renameable via `CommandWithName`), so confirm both against the project's
own posh config.

The key is a **flat map of set name to a list of YAML file paths** - not an
object with fields. The map keys are what the `config` argument accepts; the
values are merged in order to build the config sesamy receives.

Field shapes: [`foomo/sesamy/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/foomo/sesamy/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

The command needs the `onepassword` provider wired in - it is a required
argument to `NewCommand`, and it is what renders secret references in the merged
config.

**The README is not this provider's.** It is a copy of the gotsrpc one: the
title, the plugin snippet (`gotsrpc.NewCommand`) and the ownbrew package all
refer to gotsrpc. Only its config sample is sesamy's. Ignore the rest.

#### Examples

```bash
# See the effective merged config for one set, secrets rendered
posh execute sesamy config my-set

# Read-only preview before provisioning
posh execute sesamy diff web my-set

# Mutates live GTM containers; with no config name this hits EVERY set
posh execute sesamy provision all my-set
```

#### References

- [sesamy](https://github.com/foomo/sesamy-cli) - the CLI this wraps
- [Provider README](https://github.com/foomo/posh-providers/blob/main/foomo/sesamy/README.md) - only the config sample applies; the rest is gotsrpc's
