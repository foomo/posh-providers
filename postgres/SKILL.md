#### Hazards

**`restore` writes into a live database and its flags can destroy data.**
`pg_restore` applies the dump to whatever the connection flags point at, and
`--clean` drops existing objects before recreating them while `--create` creates
the target database. Nothing here confirms the target: the database is chosen by
`--dbname`/`--host`/`--port`/`--username`, none of which appear in the command
name. Confirm the connection before restoring, and treat this verb as needing
explicit approval.

**`run-cmd` and `run-file` execute arbitrary SQL** against the connected
database - including `DROP`, `TRUNCATE` and `UPDATE`. The SQL comes from the
argument or the named file, so the command line may reveal nothing about what
runs; read the file first.

**The connection target defaults to your environment, not to anything visible.**
All four connection flags are optional; unset, `psql`/`pg_dump`/`pg_restore`
fall back to `PGHOST`/`PGDATABASE`/`PGUSER` and their own defaults. So a command
with no flags is not "no database" - it is whichever one the environment
resolves to, which may be production.

**Every command is assembled into a shell line.** posh's shell helper joins the
arguments and runs them through `sh -c`, which is how `dump` redirects output
with `>`. It also means SQL, filenames and flag values are subject to shell
interpretation - a filename or `--command` containing shell metacharacters will
not behave literally.

#### Behaviour

**Compression needs the zip provider.** `CommandWithZip` is optional: without it
the `--zip-cred` flag offers no values and `dump --zip`/`--zip-cred` fails with a
message naming the option, rather than compressing. The other verbs are
unaffected.

`dump` names the file itself: `<dirname>/<database>-<YYYYMMDDhhmmss>` plus
`.sql`, or `.dump` when `--dump` is set - which also forces `--format=custom`.
The directory is created if missing (mode `0700`). You cannot choose the
filename, only the directory.

The two zip options behave differently from each other. `--zip` produces an
unencrypted archive; `--zip-cred <name>` produces a **second**, password
protected one from the same file - they are independent, so passing both creates
two archives. Neither removes the plaintext dump, which stays on disk beside the
archive. `--zip-cred` resolves a 1Password secret at run time.

Only flags you actually typed are forwarded by `dump` (`fs.Visited()`), while
`restore` forwards `r.Flags()` and the root and `run-*` leaves do too - so the
declared-but-unset flags never reach the binaries.

The root itself is a passthrough to `psql` with the arguments and flags you
gave, so `postgres` with no subcommand opens an interactive `psql` session -
which blocks until exited and is not usable unattended.

`restore` puts the filename **after** the forwarded flags and additional args,
so a stray additional arg can be interpreted as the input file by `pg_restore`.

#### Configuration

**This provider has no config key of its own** - there is no `Config` type, no
`config.schema.json` and no `config.base.json`, so there is nothing to set in
the project's posh config and no entry for it in `posh.schema.json`.

It does depend on the `arbitrary/zip` provider's configuration: `--zip-cred`
completes from that provider's `credentials` keys, so those live under the `zip`
config key, not here. See
[`arbitrary/zip/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/arbitrary/zip/config.schema.json).

The command cannot be renamed; `CommandWithZip` is the only option. `psql`,
`pg_dump` and `pg_restore` must all be on `PATH`.

**The README is not this provider's** - it is titled "POSH gotsrpc provider".
Only its dependency list applies; its plugin snippet omits `CommandWithZip`, so compression is unavailable to a project wired from it.

#### Examples

```bash
# Timestamped .sql into ./dumps, plus a password-protected archive
posh execute postgres dump mydb ./dumps --zip-cred default --dbname mydb

# Destructive: drops objects before recreating them
posh execute postgres restore ./dumps/mydb-20240101120000.sql --clean --dbname mydb

# Arbitrary SQL against whatever the connection flags resolve to
posh execute postgres run-cmd "SELECT count(*) FROM users" --dbname mydb
```

#### References

- [`pg_dump`](https://www.postgresql.org/docs/current/app-pgdump.html) / [`pg_restore`](https://www.postgresql.org/docs/current/app-pgrestore.html) / [`psql`](https://www.postgresql.org/docs/current/app-psql.html)
- [`arbitrary/zip`](https://github.com/foomo/posh-providers/blob/main/arbitrary/zip/README.md) - supplies the credentials `--zip-cred` names
- [Provider README](https://github.com/foomo/posh-providers/blob/main/postgres/README.md) - note its title says gotsrpc and its snippet omits `CommandWithZip`, so compression is not wired
