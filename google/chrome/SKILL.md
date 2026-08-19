#### Hazards

**This launches a desktop browser and needs a graphical session.** It is not
usable unattended - an agent has nowhere for the window to go, and nothing here
reports what the page did or whether Chrome even started successfully.

**The profile decides which environment the browser is pointed at, and it also
decides which stored session is used.** Each profile gets its own
`--user-data-dir`, so cookies, logins and extensions persist per profile between
runs - a profile that was logged into production stays logged in. Naming a
different profile is the only thing separating a staging session from a
production one, and the URL is taken from the config rather than shown on the
command line.

**A URL argument is passed to Chrome unvalidated.** It overrides the profile's
configured URL and is handed to the browser as typed, so whatever it points at
is opened in a session that may already hold credentials for that profile.

#### Behaviour

**The command returns immediately and does not wait for the browser.** Chrome is
started with `Start()` rather than run to completion, so posh reports success as
soon as the process spawns - a browser that fails during startup, or a binary
path that is wrong in a way only Chrome notices, still looks like success here.
The process is created from the command's context, so it is killed if that
context is cancelled.

Every launch passes `--no-default-browser-check` and `--no-first-run`, plus a
`--user-data-dir` under the configured `path` named after the profile argument.
That directory is where the profile's state accumulates; deleting it resets the
profile.

`--incognito` is added when either the top-level `incognito` or the profile's own
is set - the two are OR'd, so a global `incognito: true` cannot be overridden by
a profile.

`app: true` changes how the URL is passed: `--app=<url>` opens a chromeless
application window instead of a normal tab. With no URL from either the argument
or the profile, neither form is passed and Chrome opens its default page.

A profile's `proxy` is passed **verbatim** to Chrome as `--proxy-server`. It is
not a reference to any other provider's configuration - despite what the field's
old doc comment implied - so it must be a full proxy URL such as
`socks5://localhost:1080`. Pointing it at a tunnel that is not running gives a
browser that cannot reach anything, with no error from posh.

An unknown profile name is rejected with `profile not found: <name>` before
Chrome starts. The profile argument is required; the URL is optional.

#### Configuration

Config key `chrome` by default, overridable via `CommandWithConfigKey` (and the
command renameable via `CommandWithName`), so confirm both against the project's
own posh config.

Field shapes: [`google/chrome/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/google/chrome/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

`path` is created at startup if set, and is the parent of every profile's data
directory. `command` overrides the browser binary; left empty it defaults per
platform - the macOS and Windows defaults are absolute paths to Google Chrome,
the Linux one is `google-chrome` on `PATH`. A Chromium-based browser at another
location needs `command` set explicitly.

The `profiles` map keys are the names this command accepts - they are also used
as directory names, so they must be filesystem-safe.

#### Examples

```bash
# Opens the profile's configured url in its own persistent session
posh execute chrome staging

# Overrides the profile url, reusing that profile's cookies and logins
posh execute chrome staging https://staging.example.com/admin
```

#### References

- [Chromium command-line switches](https://peter.sh/experiments/chromium-command-line-switches/) - `--user-data-dir`, `--app`, `--proxy-server`
- [Provider README](https://github.com/foomo/posh-providers/blob/main/google/chrome/README.md) - plugin wiring and a sample config
