package beam

import (
	"sort"

	"github.com/samber/lo"
)

// Config holds the beam provider's configuration.
type Config struct {
	// Clusters reachable through a cloudflared tunnel, keyed by the name the command accepts
	Clusters map[string]Cluster `json:"clusters" yaml:"clusters"`
	// Databases reachable through a cloudflared tunnel, keyed by the name the command accepts
	Databases map[string]Database `json:"databases" yaml:"databases"`
}

// GetDatabase returns the Database configured under name, or the zero value
// if name is not configured.
func (c Config) GetDatabase(name string) Database {
	return c.Databases[name]
}

// DatabaseNames returns the configured database names, sorted
// alphabetically.
func (c Config) DatabaseNames() []string {
	ret := lo.Keys(c.Databases)
	sort.Strings(ret)

	return ret
}

// DatabaseExists reports whether name is a configured database.
func (c Config) DatabaseExists(name string) bool {
	_, ok := c.Databases[name]
	return ok
}

// GetCluster returns the Cluster configured under name, or the zero value if
// name is not configured.
func (c Config) GetCluster(name string) Cluster {
	return c.Clusters[name]
}

// ClusterNames returns the configured cluster names, sorted alphabetically.
func (c Config) ClusterNames() []string {
	ret := lo.Keys(c.Clusters)
	sort.Strings(ret)

	return ret
}

// ClusterExists reports whether name is a configured cluster.
func (c Config) ClusterExists(name string) bool {
	_, ok := c.Clusters[name]
	return ok
}
