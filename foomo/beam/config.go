package beam

import (
	"sort"

	"github.com/samber/lo"
)

type Config map[string]Tunnel

func (c Config) GetTunnel(name string) Tunnel {
	return c[name]
}

func (c Config) GetTunnelNames() []string {
	ret := lo.Keys(c)
	sort.Strings(ret)
	return ret
}
