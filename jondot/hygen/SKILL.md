#### Hazards

**This generates files into your working tree.** The selected template decides
what is written and where, and it can overwrite existing files. The command line
names a template and a target path but not the files that result - read the
template before running it, and prefer a clean working tree so `git diff` shows
what appeared.

**`--dry` is the only preview available.** It renders without saving. Use it
before the real run on any template you have not read.

Nothing validates the target path, and the template is free to write outside it.
The upstream generator is invoked as `scaffold`, so what a template emits is
governed entirely by the files under the configured template directory.

#### Behaviour

`Validate` stats `<templatePath>/<first argument>` and rejects anything that is
not a directory, so a misspelled template name fails before the generator runs.
It also requires exactly two arguments.

**`HYGEN_TMPLS` is set to the *parent* of `templatePath`**, which is deliberate:
the upstream generator expects the directory containing the generator folder,
while completion and validation work on the directories *inside* `templatePath`.
With `templatePath: .posh/scaffold` the variable becomes `.posh` and `scaffold`
is the generator name.

The target path is passed straight through and is not completed. The template
listing is cached for the session, so a template added after the shell started is
not offered until the cache is cleared. Anything after a `--` separator is
forwarded, as are the leaf's own flags.

#### Configuration

The `hygen` key's single field `templatePath` does double duty: its immediate
subdirectories are the templates completion offers, and its parent is exported as
`HYGEN_TMPLS`. What each template writes is decided inside that directory. Note
the config-key option here is `WithConfigKey`, not `CommandWithConfigKey`.

Field shapes: [`jondot/hygen/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/jondot/hygen/config.schema.json).

**The README is not this provider's** - it is titled "POSH doctl provider"; only
its config sample and plugin snippet apply here.

#### Examples

```bash
posh execute {{cmd}} my-template ./target --dry
posh execute {{cmd}} my-template ./target
```

#### References

- [hygen](https://www.hygen.io/) - templates and the `HYGEN_TMPLS` convention
- [Provider README](https://github.com/foomo/posh-providers/blob/main/jondot/hygen/README.md) - note its title says doctl
