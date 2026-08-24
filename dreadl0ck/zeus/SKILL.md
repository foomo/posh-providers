#### Hazards

This command has **no subcommand tree at all**: everything after `[path]` is
forwarded verbatim to the `zeus` binary, so the real surface is whatever build
targets the project's own configuration directory defines. Neither the catalog
nor this fragment can list them - read the scripts under the chosen `<path>`
before running one. Consequences are whatever those scripts do, which in a build
tool routinely includes deleting build output, publishing artifacts and
deploying. Only run this unattended when the specific target is known to be
read-only.

**A path that does not exist triggers a bootstrap instead of an error.** When
`[path]` is missing from disk and ends in `/zeus`, the command runs the upstream
`bootstrap` verb in the parent directory, creating a new installation there. A
mistyped path does not fail; it writes a new scaffold into whatever directory
the typo names. Completion only offers directories that already exist, so treat
a `[path]` that did not come from completion as a request to create something.

`Validate` checks only that `[path]` exists and is a directory, never that it is
the right kind of directory. An unrelated existing directory passes and the
command then runs against that directory's **parent**: `svc` validates and execs
with `-C .`, i.e. the project root, silently targeting something other than what
was named. The `/zeus` suffix is enforced only on the bootstrap branch. Pass a
path completion offered and the two branches agree.

#### Behaviour

`[path]` is the configuration **directory** itself, not the project it belongs
to - the command execs with `-C <parent-of-path>`. Completion walks the checkout
for directories named `zeus`, ignoring dotted directories and `node_modules`.

Nothing is suggested past `[path]`, so the forwarded arguments - where the target
is actually named - get no completion. A `--` separator is forwarded to the
binary as a literal argument along with everything after it. Bootstrapping clears
this command's own completion cache so new directories appear immediately.

#### Examples

```bash
posh execute {{cmd}} ./svc/zeus build
```

#### References

- [zeus](https://github.com/dreadl0ck/zeus)
- [provider README](https://github.com/foomo/posh-providers/blob/main/dreadl0ck/zeus/README.md)
