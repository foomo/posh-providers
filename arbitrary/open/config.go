package open

import (
	"github.com/foomo/posh-providers/onepassword"
)

type (
	// Config maps router names to their ConfigRouter definitions.
	Config map[string]ConfigRouter
	// ConfigRouter is a named base URL together with its nested routes.
	ConfigRouter struct {
		// Router base url
		URL string `json:"url" yaml:"url"`
		// Router Child routes
		Routes map[string]ConfigRoute `json:"routes" yaml:"routes"`
		// Router descriotion
		Description string `json:"description" yaml:"description"`
	}
	// ConfigRoute is a single route path, optionally nesting further routes and
	// a 1Password-backed basic-auth secret.
	ConfigRoute struct {
		// Route path
		Path string `json:"path" yaml:"path"`
		// Route description
		Description string `json:"description" yaml:"description"`
		// Child routes
		Routes map[string]ConfigRoute `json:"routes" yaml:"routes"`
		// Basic authentication secret
		BasicAuth *onepassword.Secret `json:"basicAuth" yaml:"basicAuth"`
	}
)

// RouteForPath resolves the route at paths, descending through nested Routes
// for every element except the last, which names the returned route.
func (c ConfigRouter) RouteForPath(paths []string) ConfigRoute {
	paths, route := paths[0:len(paths)-1], paths[len(paths)-1]
	routes := c.RoutesForPath(paths)

	return routes[route]
}

// RoutesForPath returns the route map nested at paths, descending one level
// of Routes per path element and stopping early if an element is not found.
func (c ConfigRouter) RoutesForPath(paths []string) map[string]ConfigRoute {
	routes := c.Routes
	for _, path := range paths {
		value, ok := routes[path]
		if !ok {
			break
		}

		routes = value.Routes
	}

	return routes
}
