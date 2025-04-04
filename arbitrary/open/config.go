package open

import (
	"github.com/foomo/posh-providers/onepassword"
)

type (
	Config       map[string]ConfigRouter
	ConfigRouter struct {
		// Router base url
		URL string `json:"url" yaml:"url"`
		// Router Child routes
		Routes map[string]ConfigRoute `json:"routes" yaml:"routes"`
		// Router descriotion
		Description string `json:"description" yaml:"description"`
	}
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

func (c ConfigRouter) RouteForPath(paths []string) ConfigRoute {
	paths, route := paths[0:len(paths)-1], paths[len(paths)-1]
	routes := c.RoutesForPath(paths)
	return routes[route]
}

func (c ConfigRouter) RoutesForPath(paths []string) map[string]ConfigRoute {
	routes := c.Routes
	for _, path := range paths {
		if value, ok := routes[path]; ok {
			routes = value.Routes
			break
		}
	}
	return routes
}
