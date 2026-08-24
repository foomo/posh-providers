#### Hazards

**`restore` writes into a live database and its flags can destroy data.** The dump is
applied to whatever the connection flags point at; `--clean` drops existing objects
before recreating them and `--create` creates the target database. Nothing here confirms
the target: the database is chosen by connection flags that appear nowhere in the command
name. Confirm the connection before restoring, and treat this verb as needing explicit
approval.

**`run-cmd` and `run-file` execute arbitrary SQL** against the connected database,
including `DROP` and `TRUNCATE`. The SQL comes from the argument or the named file, so the
command line may reveal nothing about what runs; read the file first.

**The connection target defaults to your environment, not to anything visible.** The
connection flags are optional; unset, the binaries fall back to
`PGHOST`/`PGDATABASE`/`PGUSER`. So a command with no flags is not "no database" — it is
whichever one the environment resolves to, possibly production.

The root is a passthrough to `psql`, so a bare `{{cmd}}` opens an **interactive session**
that blocks until exited and is unusable unattended.

#### Behaviour

**Compression needs the zip provider**, wired by an optional constructor option: without
it the zip options fail with a message naming the option rather than compressing.

`dump` names the file itself — database plus timestamp, with an extension following the
format flag. You choose the directory, not the filename.

The two zip options differ: one produces an unencrypted archive, the credential form a
**second**, password-protected one from the same file, so passing both creates two
archives. Neither removes the plaintext dump, which stays on disk beside the archive.

#### Configuration

**This provider has no config key of its own** — there is no `Config` type and no schema,
so there is nothing to set in the project's posh config.

The credential flag completes from the `arbitrary/zip` provider's credential names, under
its own key:
[`arbitrary/zip/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/arbitrary/zip/config.schema.json).

The command cannot be renamed. `psql`, `pg_dump` and `pg_restore` must be on `PATH`.

**The README is not this provider's** — it is titled "POSH gotsrpc provider", and its
snippet omits the zip option, so compression is unavailable to a project wired from it.

#### Examples

```bash
# Timestamped dump into ./dumps, plus a password-protected archive
posh execute {{cmd}} dump mydb ./dumps --zip-cred default --dbname mydb

# Destructive: drops objects before recreating them
posh execute {{cmd}} restore ./dumps/mydb-20240101120000.sql --clean --dbname mydb
```

#### References

- [`pg_dump`](https://www.postgresql.org/docs/current/app-pgdump.html) / [`pg_restore`](https://www.postgresql.org/docs/current/app-pgrestore.html) / [`psql`](https://www.postgresql.org/docs/current/app-psql.html)
