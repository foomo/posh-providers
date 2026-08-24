#### Hazards

This launches a desktop browser and needs a graphical session, so it is not
usable unattended - an agent has nowhere for the window to go, and nothing here
reports what the page did.

The profile decides both which environment is opened and which stored session is
used. Each profile gets its own `--user-data-dir`, so cookies, logins and
extensions persist between runs: a profile that was logged into production stays
logged in. Naming a different profile is the only thing separating a staging
session from a production one, and the URL comes from the config rather than the
command line - so the invocation does not show where it is pointed.

A URL argument is handed to the browser as typed, overriding the profile's
configured URL, and is opened in a session that may already hold credentials for
that profile.

#### Behaviour

The command returns as soon as the process spawns - Chrome is started, not run to
completion - so posh reports success even when the browser fails during startup
or the binary path is wrong. The process dies if the command's context is
cancelled.

The top-level `incognito` and a profile's own are OR'd, so a global
`incognito: true` cannot be overridden per profile. With no URL from either the
argument or the profile, Chrome opens its default page.

A profile's `proxy` is passed verbatim as `--proxy-server`; it is not a reference
to another provider's tunnel configuration, so it must be a full proxy URL such
as `socks5://localhost:1080`. Pointing it at a tunnel that is not running gives a
browser that reaches nothing, with no error from posh.

An unknown profile is rejected before Chrome starts.

#### Configuration

The `chrome` key holds the profile map plus the parent `path` their data
directories live under. The `profiles` keys are both the accepted argument values
and directory names, so they must be filesystem-safe; deleting one resets that
profile.

`command` overrides the browser binary; left empty it defaults per platform, so a
Chromium-based browser installed elsewhere needs it set explicitly.

Field shapes: [`google/chrome/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/google/chrome/config.schema.json).

#### Examples

```bash
# Opens the profile's configured url in its own persistent session
posh execute {{cmd}} staging

# Overrides the profile url, reusing that profile's cookies and logins
posh execute {{cmd}} staging https://staging.example.com/admin
```

#### References

- [Chromium command-line switches](https://peter.sh/experiments/chromium-command-line-switches/) - `--user-data-dir`, `--app`, `--proxy-server`
- `google/chrome/README.md` for plugin wiring and a config sample
