package beam

import (
	"sort"

	"github.com/samber/lo"
)

type Tunnel struct {
	Clusters map[string]Cluster `json:"clusters" yaml:"clusters"`
}

func (c Tunnel) GetCluster(name string) Cluster {
	return c.Clusters[name]
}

func (c Tunnel) GetClusterNames() []string {
	ret := lo.Keys(c.Clusters)
	sort.Strings(ret)

	return ret
}
