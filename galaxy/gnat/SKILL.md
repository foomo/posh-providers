#### Hazards

**This is an interactive terminal UI and it does not return.** The wrapped binary
is a full-screen TUI client for NATS JetStream; running it blocks until a human
quits and it needs a real terminal. It is not usable unattended - an agent that
runs it hangs rather than getting output, and there is no one-shot mode reachable
from this command.

**The profile decides which NATS server you are pointed at, and nothing marks
which one is production.** The name resolves to a URL from the config and is
passed as `-url`, so a local profile and a production one differ by a single
argument. JetStream state is mutable from inside the UI - streams and consumers
can be purged or deleted there - so the blast radius is whatever the selected
server allows, not whatever this command line suggests. Confirm the profile
before handing the session to anyone.

An unknown profile name *is* rejected with `given profile not found: <name>`
rather than falling through to an empty URL, and the argument is required, so
there is no "all profiles" default.

#### Behaviour

`configDir` is exported as `XDG_CONFIG_HOME` for the child process only, keeping
the tool's own config and history with the project. Left unset, the tool reads
and writes the user's real `$XDG_CONFIG_HOME`, shared with every other instance
on the machine.

Profile suggestions carry each profile's `description`, so tab completion in the
interactive shell is the reliable way to see what a name points at - the
descriptions never appear in the rendered usage block.

Everything after the profile name is forwarded to the binary, including args and
flags after a `--` separator, so upstream flags this tree does not model still
work. `-url` is always supplied from the profile and `-theme` whenever the config
sets one, so passing either yourself sends it twice.

#### Configuration

The `profiles` map under the `gnat` key holds the names this command accepts.
There is no way to pass a URL directly, so the config is the only surface for
reaching a server. `theme` and `configDir` apply to every profile. Note this
provider spells its options `WithName`/`WithConfigKey`, not `CommandWithName`.

Field shapes: [`galaxy/gnat/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/galaxy/gnat/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} local
```

#### References

- [gnat](https://github.com/galaxy-io/gnat) - the NATS JetStream terminal UI this wraps
- [Provider README](https://github.com/foomo/posh-providers/blob/main/galaxy/gnat/README.md)
