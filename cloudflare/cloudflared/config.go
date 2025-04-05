package cloudflared

import (
	"sort"

	"github.com/samber/lo"
)

type Config struct {
	Path   string            `json:"path" yaml:"path"`
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
