#### Hazards

**This command writes HTML files, overwriting whatever is already there.** For
each `*.mjml` source it derives an output path and compiles onto it with no
diff, confirmation or backup. Hand-edits to a generated `.html` are lost.

**Omitting the path compiles every template in the project.** Given a path, only
`*.mjml` files under a `/src/` segment below it are compiled; omitted, the whole
checkout is walked. The wide form is the default, so check the argument first.

**Compiles run concurrently and failures are not isolated.** `--parallel N` runs
N at once (default: one at a time). The group shares one context, so the first
failure cancels the rest - leaving some outputs written, some stale and some
never attempted. Rerunning after a fix is the only way back to a consistent
state.

#### Behaviour

Only sources under a `/src/` path segment are compiled - a template outside such
a directory is silently skipped and nothing reports it.

The output path is derived by string substitution, not by a suffix or segment
rule: `.mjml` becomes `.html` and `/src/` becomes `/html/`, each applied to
**every** occurrence. For the intended layout (`emails/src/welcome.mjml` to
`emails/html/welcome.html`) that is right, but a repeated segment is mangled -
`src/a/src/b.mjml` writes to `src/a/html/b.html` - and a directory whose own
name contains `.mjml` is rewritten too. The `/html/` directory is not created
here; the compiler must be able to write there.

Validation is stricter than execution: the argument must be an existing
**directory** and only one may be given, or you get `invalid [path] parameter` /
`too many arguments`.

Path suggestions are cached for the session, so a template added after the shell
started is not offered until the cache is cleared.

Anything after a `--` separator is appended to each compiler invocation, which
is how upstream options are reached; `--parallel` is consumed by the provider
and never forwarded.

#### Configuration

No config key, no schema - nothing to set in the project's posh config. What
gets compiled is decided entirely by the `/src/` convention on disk.

The command can be renamed with `CommandWithName`. There is no option to swap
the binary invocation, so it must be on `PATH`.

#### References

- [MJML documentation](https://documentation.mjml.io/) - the compiler and its options
- [Provider README](https://github.com/foomo/posh-providers/blob/main/mjmlio/mjml/README.md)
