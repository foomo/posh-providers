#### Hazards

**This command writes HTML files, overwriting whatever is already there.** For
each `*.mjml` source it derives an output path and compiles onto it with no
diff, confirmation or backup. Hand-edits to a generated `.html` are lost.

**The `path` argument scopes the run.** Given one, only `*.mjml` files under a
`/src/` directory below that path are compiled; omitted, the whole project is
walked. So a bare `mjml` rewrites every template in the repo — check the argument
before running it, since the wide form is the default.

**Compiles run concurrently and failures are not isolated.** `--parallel N`
runs N at once (default: one at a time). The group uses a shared context, so the
first failure cancels the rest - leaving some outputs written, some stale and
some never attempted. Rerunning after a fix is the only way back to a consistent
state.

#### Behaviour

Only sources under a `/src/` path segment are compiled. The lookup finds every
`*.mjml` in the project and then keeps only those whose path contains `/src/`,
so a template outside such a directory is silently skipped - it will not be
compiled and nothing reports it.

The output path is derived by string substitution on the source path, not by a
suffix or segment rule: `.mjml` becomes `.html` and `/src/` becomes `/html/`,
each applied to **every** occurrence in the path. For the intended layout
(`emails/src/welcome.mjml` → `emails/html/welcome.html`) that is right, but a
path with a repeated segment is mangled - `src/a/src/b.mjml` writes to
`src/a/html/b.html`, and a directory whose own name contains `.mjml` is
rewritten too. The `/html/` directory is not created by this provider; mjml must
be able to write there.

Validation is stricter than execution: the argument must be an existing
**directory** and only one may be given, or you get `invalid [path] parameter` /
`too many arguments`. Passing none is valid and behaves identically to passing
one, per the hazard above.

The `path` suggestions are directories that contain a `/src/` with `*.mjml`
files under them, discovered once and cached for the session - a template added
after the shell started is not offered, and not compiled, until the cache is
cleared.

Anything after a `--` separator is appended to each `mjml` invocation, which is
how upstream options are reached; this tree declares only `--parallel`, which is
consumed by the provider and never forwarded.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type, no
`config.schema.json` and no `config.base.json`, so there is nothing to set in
the project's posh config and no entry for it in `posh.schema.json`. What gets
compiled is decided entirely by the `/src/` convention on disk.

The command can be renamed with `CommandWithName`. There is no option to swap
the `mjml` binary invocation, so it must be on `PATH`.

#### Examples

```bash
# Compiles every *.mjml under a /src/ directory in the project
posh execute mjml

# Identical in effect - the path is reported, not applied
posh execute mjml emails/newsletter

# Four at a time, passing an upstream flag through
posh execute mjml --parallel 4 -- --config.minify true
```

#### References

- [MJML documentation](https://documentation.mjml.io/) - the compiler and its options
- [Provider README](https://github.com/foomo/posh-providers/blob/main/mjmlio/mjml/README.md) - note it describes `path` as scoping the run, which it does not
