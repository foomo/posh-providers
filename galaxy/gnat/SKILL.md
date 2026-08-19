#### Hazards

**This is an interactive terminal UI and it does not return.** `gnat` is a
full-screen TUI client for NATS JetStream; running it blocks until a human quits
the program, and it needs a real terminal. It is not usable unattended - an agent
that runs it will hang rather than get output. There is no non-interactive or
one-shot mode reachable from this command.

**The profile decides which NATS server you are pointed at, and nothing marks
which one is production.** The name resolves to a URL from the config and is
passed as `-url`; a local profile and a production one differ by a single
argument. JetStream state is mutable from inside the UI - streams and consumers
can be purged or deleted there - so the blast radius is whatever the selected
server allows, not whatever this command line suggests. Confirm the profile
before handing the session to anyone.

An unknown profile name *is* rejected: `Profile` returns
`given profile not found: <name>` rather than falling through to an empty URL.
The argument is required, so there is no "all profiles" default to worry about.

#### Behaviour

**`configDir` currently has no effect.** It is meant to set `$XDG_CONFIG_HOME`
for the gnat process so its files stay with the project, and the config comment
and README both say so. But `execute` appends it with `.Args(...)` rather than
`.Env(...)` (`command.go:138`), and `Args` appends to argv while `Env` sets
`cmd.Env` - so `XDG_CONFIG_HOME=<dir>` is passed to `gnat` as a trailing
positional argument and the environment variable is never set. gnat therefore
reads and writes its config under the user's real `$XDG_CONFIG_HOME`. The
sibling `galaxy/tempo` does the same thing correctly with `.Env(envs...)`.

The profile suggestions carry each profile's `description`, so tab completion in
the interactive shell is the reliable way to see what a name points at - the
descriptions do not appear in the rendered usage block.

Everything after the profile name is forwarded to the `gnat` binary: the
command's own flags, plus any additional args and flags after a `--` separator.
Upstream flags this tree does not model still work, and inherit upstream
behaviour. `-url` is always supplied from the profile, and `-theme` whenever the
config sets one, so passing either yourself means gnat sees it twice.

#### Configuration

Config key `gnat` by default, overridable via `WithConfigKey` (and the command
renameable via `WithName` - note this provider spells it `WithName`, not the
`CommandWithName` most others use), so confirm both against the project's own
posh config.

Field shapes: [`galaxy/gnat/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/galaxy/gnat/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

The `profiles` map keys are the names this command accepts; there is no way to
pass a URL directly, so the config is the only surface for reaching a server.
`theme` and `configDir` are optional and apply to every profile.

#### Examples

```bash
# Opens the TUI against the named profile - blocks until quit, needs a terminal
posh execute gnat local
```

#### References

- [gnat](https://github.com/galaxy-io/gnat) - the NATS JetStream terminal UI this wraps
- [Provider README](https://github.com/foomo/posh-providers/blob/main/galaxy/gnat/README.md)
