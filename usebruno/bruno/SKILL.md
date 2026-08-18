#### Hazards

`run` makes real HTTP calls against whichever environment is named, so it is
never read-only. Check the environment before running: the same request set
points at local or production depending on that one argument.

Naming no request runs the entire collection. Name requests to run a subset.

`env` renders the collection's `bruno.env` template to `.env` via
`op inject`, overwriting it. It needs a 1Password provider wired into the
command (`bruno.CommandWithOnePassword`) and fails with "you must provide a
one-password configuration" otherwise. Run it before `run` when the environment
references `process.env` secrets.

`open` launches the desktop app and needs a graphical session. It is not usable
under an agent.

#### Behaviour

The environment and request names are not fixed. They are globbed off disk from
the collection directory in this checkout - environments from
`environments/*.bru`, requests from the `*.bru` files outside it - so they change
with the branch and cannot be listed in this document. Run `bruno list` first;
what it prints is exactly what `run` accepts.

#### Configuration

The collection path comes from the `bruno` key of this project's posh config -
read it to find where the `.bru` files live. `bruno` is only the default key,
overridable via `CommandWithConfigKey`, so confirm the actual one against the
project's own file.

Field shapes: [`usebruno/bruno/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/usebruno/bruno/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same
schema is bundled under in the project's `posh.schema.json`.

#### Examples

```bash
posh execute bruno list
posh execute bruno run local
posh execute bruno run local auth/login.bru --bail
```

#### References

- https://docs.usebruno.com/bru-cli/overview
- `usebruno/bruno/README.md` for plugin wiring and the 1Password setup
