#### Hazards

**This generates files into your working tree.** hygen writes whatever the
selected template directory produces, at paths the template itself decides, and
it can overwrite existing files. The command line names a template and a target
path but not the files that result - read the template before running it, and
prefer a clean working tree so `git diff` shows what appeared.

**`--dry` is the safe way to look first.** It makes hygen render without saving,
which is the only preview available here. It is declared on the `template` leaf
only.

**The template argument is never validated as a template.** `Validate` checks
`<templatePath>/<first argument>`, but at this leaf the first argument is the
literal word `template` - the node's own name - not the template you named. So
validation asks whether a directory called `template` exists, and the real
template name is checked by nothing on the posh side. A misspelled template
reaches hygen unverified.

#### Behaviour

**The only runnable form is `hygen template <path>`.** The root node declares a
`path` argument but has no `Execute`, so that argument is unreachable - the sole
executable leaf is `template`, whose completion offers the directories found
under the configured `templatePath`.

**The literal word `template` is forwarded to hygen.** `execute` passes
`r.Args()` verbatim after `hygen scaffold`, and at this leaf those args begin
with the node name, so the real invocation is
`hygen scaffold template <path> ...`. Whether that is intended depends on the
template layout: hygen's own convention is `hygen <generator> <action>`, so this
provider is effectively pinned to a generator named `scaffold` with an action
named `template`.

**`HYGEN_TMPLS` is set to the *parent* of `templatePath`.** With
`templatePath: .posh/scaffold` the environment variable becomes `.posh` - while
completion lists directories *inside* `.posh/scaffold` and validation stats
paths inside it too. The three disagree about what the template root is, which
is worth checking against your own layout before trusting completion.

The second `path` argument is passed straight through and is deliberately not
completed - its `Suggest` returns nothing.

The template directory listing is cached for the session, so a template added
after the shell started is not offered until the cache is cleared.

Anything after a `--` separator is forwarded to hygen, as are the leaf's own
flags.

#### Configuration

Config key `hygen` by default, overridable via `WithConfigKey` (and the command
renameable via `CommandWithName`) - note the config-key option here is
`WithConfigKey`, not the `CommandWithConfigKey` some siblings use.

Field shapes: [`jondot/hygen/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/jondot/hygen/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

The single field `templatePath` does double duty: its immediate subdirectories
are the templates offered by completion, and its **parent** is exported to hygen
as `HYGEN_TMPLS`. What each template actually writes is decided inside that
directory, not by anything here.

**The README is not this provider's.** It is titled "POSH doctl provider"; only
its config sample and plugin snippet apply to hygen.

#### Examples

```bash
# Render without saving - the only preview available
posh execute hygen template ./target --dry

# Writes files into the working tree
posh execute hygen template ./target
```

#### References

- [hygen](https://www.hygen.io/) - the generator this wraps, and the `HYGEN_TMPLS` convention
- [Provider README](https://github.com/foomo/posh-providers/blob/main/jondot/hygen/README.md) - note its title says doctl
