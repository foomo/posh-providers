package cloudflared

import (
	"sort"

	"github.com/samber/lo"
)

type Config struct {
	// Path used as HOME for every cloudflared call, keeping its credentials per project
	Path string `json:"path" yaml:"path"`
	// Access configurations, keyed by the name given to `access connect`
	Access map[string]Access `json:"access" yaml:"access"`
}

func (c Config) AccessNames() []string {
	ret := lo.Keys(c.Access)
	sort.Strings(ret)

	return ret
}

func (c Config) GetAccesss(name string) Access {
	return c.Access[name]
}
