package open

import (
	"github.com/foomo/posh-providers/onepassword"
)

type (
	Config       map[string]ConfigRouter
	ConfigRouter struct {
		// Router base url
		URL string `yaml:"url"`
		// Router Child routes
		Routes map[string]ConfigRoute `yaml:"routes"`
		// Router descriotion
		Description string `yaml:"description"`
	}
	ConfigRoute struct {
		// Route path
		Path string `yaml:"path"`
		// Route description
		Description string `yaml:"description"`
		// Child routes
		Routes map[string]ConfigRoute `yaml:"routes"`
		// Basic authentication secret
		BasicAuth *onepassword.Secret `yaml:"basicAuth"`
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
