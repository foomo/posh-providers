#### Hazards

**This is an interactive terminal UI and it does not return.** `tempo` is a
full-screen TUI client for Temporal; running it blocks until a human quits the
program, and it needs a real terminal. It is not usable unattended - an agent
that runs it will hang rather than get output. There is no non-interactive or
one-shot mode reachable from this command.

**The profile decides which Temporal server you are pointed at, and nothing
marks which one is production.** The name resolves to an address from the config
and is passed as `--address`; a local profile and a production one differ by a
single argument. Workflow state is mutable from inside the UI - running
workflows can be terminated or cancelled and signals sent there - so the blast
radius is whatever the selected server allows, not whatever this command line
suggests. Confirm the profile before handing the session to anyone.

An unknown profile name *is* rejected: `Profile` returns
`given profile not found: <name>` rather than falling through to an empty
address. The argument is required, so there is no "all profiles" default to
worry about.

#### Behaviour

`--address` takes a `host:port`, not a URL, despite the config field being
called `url` - the sample config uses `localhost:7233`. A scheme-prefixed value
is passed through unchanged and fails at the Temporal client rather than here.

The profile suggestions carry each profile's `description`, so tab completion in
the interactive shell is the reliable way to see what a name points at - the
descriptions do not appear in the rendered usage block.

Everything after the profile name is forwarded to the `tempo` binary: the
command's own flags, plus any additional args and flags after a `--` separator.
Upstream flags this tree does not model still work, and inherit upstream
behaviour. `--address` is always supplied from the profile, and `--theme`
whenever the config sets one, so passing either yourself means tempo sees it
twice.

`configDir` is applied as a real environment variable (`XDG_CONFIG_HOME`) for
the tempo process, so its config files stay with the project rather than in the
user's home directory. The sibling `galaxy/gnat` has the same option and behaves
the same way.

#### Configuration

Config key `tempo` by default, overridable via `WithConfigKey` (and the command
renameable via `WithName` - note this provider spells it `WithName`, not the
`CommandWithName` most others use), so confirm both against the project's own
posh config.

Field shapes: [`galaxy/tempo/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/galaxy/tempo/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

The `profiles` map keys are the names this command accepts; there is no way to
pass an address directly, so the config is the only surface for reaching a
server. `theme` and `configDir` are optional and apply to every profile.

#### Examples

```bash
# Opens the TUI against the named profile - blocks until quit, needs a terminal
posh execute tempo local
```

#### References

- [tempo](https://github.com/galaxy-io/tempo) - the Temporal terminal UI this wraps
- [Provider README](https://github.com/foomo/posh-providers/blob/main/galaxy/tempo/README.md)
