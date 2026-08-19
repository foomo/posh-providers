#### Hazards

**This launches a browser and needs a graphical session.** It resolves a URL and
hands it to the system opener, so it is not usable unattended - an agent has
nowhere for the window to go, and nothing here reports what the page did.

**A route may embed live credentials in the URL it opens.** When the selected
route sets `basicAuth`, the username and password are fetched from 1Password on
each invocation and written into the URL as `https://user:pass@host/...`. That
requires a signed-in 1Password session and may prompt. It also means the
credentials travel wherever that URL goes: browser history, the page's referrer
chain, and any process that can see the opener's argument list. Prefer routes
without `basicAuth` when the target does not need it.

**The router argument decides which environment you land on.** A local router
and a production one differ by a single word, and the URL is assembled from the
config rather than shown in the command line - `open admin users` says nothing
about which host it reaches. Read the `open` config key before opening a route
you do not recognise.

#### Behaviour

**Routes nested more than two levels deep are unreachable.** The lookup walks the
route tree but `break`s after the first match instead of descending
(`config.go`, `RoutesForPath`), so it only ever goes one level down. A route at
`<router> <a> <b>` resolves; one at `<router> <a> <b> <c>` yields an empty path
and is rejected by validation with `invalid [route] argument`. The failure is
safe - it never opens a wrong URL - but a deeply nested config simply cannot be
reached from here, and tab completion at that depth shows nothing. Flatten the
config if you need those routes.

An unknown router or route is rejected before anything opens: `invalid [router]
argument` / `invalid [route] argument`. Both arguments are required.

The route argument is repeatable and each word descends one level, so the
completion at each position lists the child routes of what you have typed so
far - the descriptions in those suggestions come from the config and do not
appear in the rendered usage block.

**A `basicAuth` secret's `field` is ignored.** The provider reads the same
1Password item twice, overriding the field with `username` and then `password`,
so whatever `field` the config sets has no effect - the item must expose fields
under exactly those names.

The final URL is `router.url + route.path` concatenated as strings, with no
separator inserted, so the config must supply the leading `/` on each path.

#### Configuration

Config key `open` by default, overridable via `WithConfigKey` (and the command
renameable via `CommandWithName`) - note the config-key option here is
`WithConfigKey`, not the `CommandWithConfigKey` some siblings use. Confirm both
against the project's own posh config.

The key is a **map of router name to router**, each with a base `url` and a tree
of named `routes`; a route has a `path`, an optional `description`, optional
child `routes`, and an optional `basicAuth` 1Password secret reference.

Field shapes: [`arbitrary/open/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/arbitrary/open/config.schema.json).
That URL is also the schema's `$id`, so it is the `$defs` key the same schema is
bundled under in the project's `posh.schema.json`.

The config is the only surface - there is no way to pass a URL directly. The
command needs a `onepassword` provider wired into `NewCommand`, even for routes
that use no credentials.

#### Examples

```bash
# Opens router.url + route.path in the browser
posh execute open my-router my-route

# Nested one level - the deepest reachable form
posh execute open my-router admin users
```

#### References

- [Provider README](https://github.com/foomo/posh-providers/blob/main/arbitrary/open/README.md) - plugin wiring and a sample config
