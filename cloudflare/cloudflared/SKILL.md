#### Hazards

`access disconnect` kills any `cloudflared` process on the machine whose command
line mentions the configured hostname - process discovery is a system-wide scan,
not a registry of what this shell started. It can therefore terminate a tunnel a
different posh session or the user started by hand.

The `tunnel` subcommands act on the Cloudflare account, not on this machine:
`create`, `delete` and `route dns` mutate remote state, and a deleted tunnel
cannot be recovered. `tunnel create` also prints the new tunnel's decoded token
to the log, so its output carries a credential.

`tunnel login` opens a browser to complete authentication and cannot finish
without a graphical session. Ask a human to run it if the other tunnel
subcommands fail on authentication.

#### Behaviour

`access list` scans every process named `cloudflared`, so it shows tunnels this
shell never started; the reverse also holds, and an empty list only means no such
process exists anywhere. `access connect` starts a detached background process in
its own process group that outlives the command, and refuses with
"connection already exists" rather than starting a second one for the same
hostname. `tunnel list` and `tunnel info` are read-only.

Every invocation runs with `HOME` pointed at this provider's configured path, so
`cloudflared` credentials are kept per project rather than in the user's real home
directory. That directory is created on startup.

#### Configuration

Access names are fixed for the checkout - they come from the `cloudflared` key of
this project's posh config, not a live lookup, so reading that key tells you
every valid name plus the local port each binds on `127.0.0.1`.

Field shapes: [`cloudflare/cloudflared/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/cloudflare/cloudflared/config.schema.json).

#### Examples

```bash
posh execute {{cmd}} access list
posh execute {{cmd}} access connect <name>
posh execute {{cmd}} tunnel list
```

#### References

- https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/
- `cloudflare/cloudflared/README.md` for plugin wiring and sample config
