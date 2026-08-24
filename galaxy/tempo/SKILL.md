#### Hazards

**This is an interactive terminal UI and it does not return.** The wrapped binary
is a full-screen TUI client for Temporal; running it blocks until a human quits
and it needs a real terminal. It is not usable unattended - an agent that runs it
hangs rather than getting output, and there is no one-shot mode reachable from
this command.

**The profile decides which Temporal server you are pointed at, and nothing marks
which one is production.** The name resolves to an address from the config and is
passed as `--address`, so a local profile and a production one differ by a single
argument. Workflow state is mutable from inside the UI - running workflows can be
terminated or cancelled and signals sent there - so the blast radius is whatever
the selected server allows, not whatever this command line suggests. Confirm the
profile before handing the session to anyone.

An unknown profile name *is* rejected with `given profile not found: <name>`
rather than falling through to an empty address, and the argument is required, so
there is no "all profiles" default.

#### Behaviour

`--address` takes a `host:port`, not a URL, despite the config field being called
`url` - the sample config uses `localhost:7233`. A scheme-prefixed value is
passed through unchanged and fails at the Temporal client rather than here.

`configDir` is exported as `XDG_CONFIG_HOME` for the child process only, keeping
the tool's own config files with the project rather than the user's home
directory. The sibling `gnat` command has the same option and behaves the same
way.

Profile suggestions carry each profile's `description`, so tab completion is the
reliable way to see what a name points at - the descriptions never appear in the
rendered usage block. Everything after the profile name is forwarded to the
binary, including args and flags after a `--` separator. `--address` is always
supplied from the profile and `--theme` whenever the config sets one, so passing
either yourself sends it twice.

#### Configuration

The `profiles` map under the `tempo` key holds the names this command accepts.
There is no way to pass an address directly, so the config is the only surface
for reaching a server. Note this provider spells its options
`WithName`/`WithConfigKey`, not `CommandWithName`.

Field shapes: [`galaxy/tempo/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/galaxy/tempo/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} local
```

#### References

- [tempo](https://github.com/galaxy-io/tempo) - the Temporal terminal UI this wraps
- [Provider README](https://github.com/foomo/posh-providers/blob/main/galaxy/tempo/README.md)
