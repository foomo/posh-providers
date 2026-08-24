#### Hazards

**This launches a browser and needs a graphical session.** It resolves a URL and
hands it to the system opener, so it is not usable unattended - nothing here
reports what the page did.

**A route may embed live credentials in the URL it opens.** When the route sets
`basicAuth`, the username and password are fetched from 1Password on each
invocation and written into the URL as `https://user:pass@host/...`. That needs a
signed-in 1Password session and may prompt. The credentials then travel wherever
that URL goes: browser history, the page's referrer chain, and any process that
can see the opener's argument list.

**The router argument decides which environment you land on.** A local router and
a production one differ by a single word, and the URL is assembled from config
rather than shown in the command line, so an invocation says nothing about which
host it reaches. Read the `open` key before opening a route you do not recognise.

#### Behaviour

Routes nest to any depth: each argument after the router descends one level of
the config's `routes` tree, and completion offers that level's children with
their configured descriptions. Both arguments are required, and an unknown one
is rejected before anything opens.

**A `basicAuth` secret's `field` is ignored.** The provider reads the same
1Password item twice, overriding the field with `username` and then `password`,
so whatever `field` the config sets has no effect - the item must expose fields
under exactly those names.

The opened URL is `router.url` concatenated with the **leaf** route's `path`, as
plain strings with no separator inserted. Ancestor paths do not accumulate, so a
nested route's `path` must be the full path from the router root, including its
leading `/`.

#### Configuration

The `open` key is a map of router name to router, each with a base `url` and a
tree of named `routes`; a route has a `path`, an optional `description`, optional
child `routes`, and an optional `basicAuth` 1Password secret reference. The
config-key option here is `WithConfigKey`, not the `CommandWithConfigKey` some
siblings use.

Field shapes: [`arbitrary/open/config.schema.json`](https://raw.githubusercontent.com/foomo/posh-providers/main/arbitrary/open/config.schema.json).

The config is the only surface - no URL can be passed directly. A `onepassword`
provider must be wired into `NewCommand` even for routes using no credentials.

#### Examples

```bash
posh execute {{cmd}} my-router my-route
posh execute {{cmd}} my-router admin users detail
```

#### References

- [Provider README](https://github.com/foomo/posh-providers/blob/main/arbitrary/open/README.md) - plugin wiring and a sample config
