#### Hazards

Every verb except `version` writes to the database named by the first argument.
`version` is the only read-only one, so it is the safe orientation step: run it
first to see where the database stands and whether it is dirty.

`drop` deletes everything in the database and `down` rolls back **every**
migration, not one - it calls `Down()`, while `down-by-one` is the single-step
verb. Both are irreversible without a backup, neither prompts, and posh does not
gate them: the first argument is the only thing separating a local database from
a production one. Do not run either unattended - ask for explicit confirmation of
the target database first.

`force` rewrites the recorded version without running any migration. It is the
recovery path for a dirty database and will happily record a version whose
migrations never ran, so the schema and the ledger silently disagree afterwards.

Database URLs may carry 1Password secret references, resolved at execution time
only when a 1Password provider is wired in via `migrate.CommandWithOnePassword`.
Without it the raw config value goes to the driver as-is and fails on an
unresolved reference.

Cancelling triggers a graceful stop rather than an immediate kill: the in-flight
migration finishes before the process exits, so it can take a moment to return
and should not be force-killed.

#### Behaviour

Argument order is `<database> <source> <verb>` - the database comes before the
migration set, and both before the thing to do. `force` and `migrate` take a
trailing version number; the rest take none.

Database and source names are not fixed. They are the keys of the `databases`
and `sources` maps in this project's posh config, so they differ per project. An
unknown name is not an error - it resolves to an empty URL and the command fails
later when the driver rejects it.

Drivers are compiled into the host posh shell by blank import, not selected at
runtime, so a URL scheme the project did not import fails even with correct
config.

#### Configuration

Read the `migrate` key for the available database and source names.

Field shapes: [`golang-migrate/migrate/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/golang-migrate/migrate/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} local default version
posh execute {{cmd}} local default up
posh execute {{cmd}} local default force 2
```

#### References

- https://github.com/golang-migrate/migrate
- `golang-migrate/migrate/README.md` for plugin wiring and the driver imports
