package cloudflared

import (
	"sort"

	"github.com/samber/lo"
)

// Config holds the cloudflared HOME directory and the named Access targets
// available to the cloudflared command.
type Config struct {
	// Path used as HOME for every cloudflared call, keeping its credentials per project
	Path string `json:"path" yaml:"path"`
	// Access configurations, keyed by the name given to `access connect`
	Access map[string]Access `json:"access" yaml:"access"`
}

// AccessNames returns the configured access names, sorted.
func (c Config) AccessNames() []string {
	ret := lo.Keys(c.Access)
	sort.Strings(ret)

	return ret
}

// GetAccesss returns the named access configuration, or the zero Access if
// name is not configured.
func (c Config) GetAccesss(name string) Access {
	return c.Access[name]
}
