package doctl

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	ConfigPath string             `json:"configPath" yaml:"configPath"`
	Clusters   map[string]Cluster `json:"clusters" yaml:"clusters"`
}

func (c Config) Cluster(name string) (Cluster, error) {
	value, ok := c.Clusters[name]
	if !ok {
		return Cluster{}, errors.Errorf("given cluster not found: %s", name)
	}

	return value, nil
}

func (c Config) ClusterNames() []string {
	return lo.Keys(c.Clusters)
}
