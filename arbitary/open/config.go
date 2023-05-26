package open

import (
	"github.com/foomo/posh-providers/onepassword"
)

type (
	Config       map[string]ConfigRouter
	ConfigRouter struct {
		URL         string                 `yaml:"url"`
		Routes      map[string]ConfigRoute `yaml:"routes"`
		Description string                 `yaml:"description"`
	}
	ConfigRoute struct {
		Path        string              `yaml:"path"`
		Description string              `yaml:"description"`
		BasicAuth   *onepassword.Secret `yaml:"basicAuth"`
	}
)
