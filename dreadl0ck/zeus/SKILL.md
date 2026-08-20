#### Hazards

This command is a **passthrough with no subcommand tree at all**: everything
after `[path]` is forwarded verbatim to the `zeus` binary, so the real surface is
whatever build targets the project's own `zeus` directory defines. The catalog
cannot list them and neither can this fragment — read the scripts under the
chosen `<path>` to find out what a target does before running it. Consequences
are whatever those scripts do, which in a build tool routinely includes deleting
build output, publishing artifacts and running deploys.

**A path that does not exist triggers a bootstrap instead of an error.** When
`[path]` is missing from disk and ends in `/zeus`, the command runs
`zeus bootstrap` in the parent directory, creating a new zeus installation
there. So a mistyped path does not fail — it writes a new scaffold into whatever
directory the typo names. Completion only ever offers directories that already
exist, so the bootstrap branch is reachable only by typing, and an agent should
treat a `[path]` it did not get from completion as a request to create something.

Only run this unattended when the specific target is known to be read-only.

#### Behaviour

`[path]` is the zeus **directory** itself, not the project it belongs to — the
command execs `zeus -C <parent-of-path>`. Completion reflects that, walking the
checkout for directories named `zeus` (ignoring dotted directories and
`node_modules`) and offering e.g. `./svc/zeus`.

`Validate` only checks that `[path]` exists and is a directory; it does not check
that it *is* a zeus directory. So an unrelated existing directory passes and the
command runs zeus against that directory's **parent**: `zeus svc` validates and
execs `zeus -C .`, i.e. the project root, silently targeting something other than
what was named. The `/zeus` suffix is only enforced on the non-existent-path
branch. Pass a path that completion offered, and the two branches agree.

Nothing is suggested past `[path]` — the forwarded arguments, which is where the
target is actually named, get no completion. A `--` separator is forwarded to
`zeus` as a literal argument along with everything after it.

Bootstrapping clears this command's completion cache, so newly created zeus
directories show up immediately; other cached lookups are unaffected.

#### Examples

```bash
# Run a target in an existing zeus directory (path as completion offers it)
posh execute zeus ./svc/zeus build

# Bootstrap a new zeus installation under svc/ - creates files
posh execute zeus ./newsvc/zeus
```

#### References

- [zeus](https://github.com/dreadl0ck/zeus)
- [provider README](https://github.com/foomo/posh-providers/blob/main/dreadl0ck/zeus/README.md)
