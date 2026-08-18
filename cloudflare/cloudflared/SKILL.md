#### Hazards

`access list` reports only the tunnel processes started by this posh shell, not
every cloudflared process on the machine. It is empty on a fresh session, so an
empty list means nothing has been connected yet - not that a tunnel failed.

`access connect` starts a background process that keeps running after the
command returns; `access disconnect` stops it. Connecting twice to the same
target is not useful, so check `access list` first.

The `tunnel` subcommands act on the Cloudflare account, not on this machine:
`create`, `delete` and `route dns` mutate remote state, and a deleted tunnel
cannot be recovered. `list` and `info` are read-only and safe for orientation.
`tunnel create` prints the new tunnel's decoded token to the log, so its output
carries a credential.

`tunnel login` opens a browser to complete authentication and cannot finish
without a graphical session. Ask a human to run it if the other tunnel
subcommands fail on authentication.

Every invocation runs with `HOME` pointed at this provider's configured path, so
cloudflared credentials are kept per project rather than in the user's real home
directory.

#### Configuration

The access names are fixed for the checkout - they come from the `cloudflared`
key of this project's posh config, not from a live lookup, so reading that key
tells you every valid name. `cloudflared` is only the default key, overridable via
`WithConfigKey`, so confirm the actual one against the project's own file.

Field shapes: [`cloudflare/cloudflared/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/cloudflare/cloudflared/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same
schema is bundled under in the project's `posh.schema.json`.

#### Examples

```bash
posh execute cloudflared access list
posh execute cloudflared access connect <name>
posh execute cloudflared tunnel list
```

#### References

- https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/
- `cloudflare/cloudflared/README.md` for plugin wiring and sample config
