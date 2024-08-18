package beam

import (
	"sort"

	"github.com/samber/lo"
)

type Config struct {
	Clusters  map[string]Cluster  `yaml:"clusters"`
	Databases map[string]Database `yaml:"databases"`
}

func (c Config) GetDatabase(name string) Database {
	return c.Databases[name]
}

func (c Config) DatabaseNames() []string {
	ret := lo.Keys(c.Databases)
	sort.Strings(ret)
	return ret
}

func (c Config) DatabaseExists(name string) bool {
	_, ok := c.Databases[name]
	return ok
}

func (c Config) GetCluster(name string) Cluster {
	return c.Clusters[name]
}

func (c Config) ClusterNames() []string {
	ret := lo.Keys(c.Clusters)
	sort.Strings(ret)
	return ret
}

func (c Config) ClusterExists(name string) bool {
	_, ok := c.Clusters[name]
	return ok
}
