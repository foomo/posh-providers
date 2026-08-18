#### Hazards

Every verb except `version` writes to the database named by the first argument.
`version` is the only read-only verb, so it is the safe one for orientation: run
it first to see where the database actually stands.

`drop` deletes everything in the database and `down` rolls back every migration.
Both are irreversible without a backup, neither prompts for confirmation, and
posh does not gate them - the first argument is the only thing separating a
local database from a production one. Do not run either unattended: ask for
explicit confirmation of the target database first. The same applies to `force`,
which rewrites the recorded version without running anything.

The argument order is `<database> <source> <verb>` - the database comes before
the migration set, and both come before the thing to do. `force` and `migrate`
take a trailing version number; the rest take none.

`force` sets the recorded version without running any migration. It is the
recovery path for a database the `version` verb reports as dirty, and it will
happily record a version whose migrations never ran.

The database and source names are not fixed. They are the keys of the `databases`
and `sources` maps in this project's posh config, so they differ per project and
cannot be listed here. Read that config to find the real names; tab completion
in the interactive shell suggests them from the same maps. An unknown name is
not an error - it resolves to an empty URL and the command fails when the driver
rejects it.

Database URLs may carry 1Password secret references. They are resolved at
execution time only when a 1Password provider is wired into the command
(`migrate.CommandWithOnePassword`); without it the raw config value is passed to
the driver as-is, which fails if it still contains an unresolved reference.

Cancelling the command triggers a graceful stop rather than an immediate kill:
the in-flight migration is allowed to finish before the process exits, so it can
take a moment to return and should not be force-killed.

The database and source drivers are compiled into the host posh shell by blank
import, not selected at runtime. A URL scheme the project did not import fails
at startup regardless of the config being correct.

#### Configuration

Read the `migrate` key of this project's posh config for the available database
and source names. Field shapes are in the schema,
`https://github.com/foomo/posh-providers/golang-migrate/migrate`.

#### Examples

```bash
posh execute migrate local default version
posh execute migrate local default up
posh execute migrate local default migrate 3
posh execute migrate local default force 2
```

#### References

- https://github.com/golang-migrate/migrate
- `golang-migrate/migrate/README.md` for plugin wiring and the driver imports
