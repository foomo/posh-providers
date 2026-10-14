package beam

import (
	"sort"

	"github.com/samber/lo"
)

// Tunnel holds the clusters reachable through a cloudflared tunnel.
type Tunnel struct {
	Clusters map[string]Cluster `json:"clusters" yaml:"clusters"`
}

// GetCluster returns the Cluster configured under name, or the zero value if
// name is not configured.
func (c Tunnel) GetCluster(name string) Cluster {
	return c.Clusters[name]
}

// GetClusterNames returns the configured cluster names, sorted
// alphabetically.
func (c Tunnel) GetClusterNames() []string {
	ret := lo.Keys(c.Clusters)
	sort.Strings(ret)

	return ret
}
