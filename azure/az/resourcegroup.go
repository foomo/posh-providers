package az

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type ResourceGroup struct {
	Name         string             `json:"name" yaml:"name"`
	Subscription string             `json:"subscription" yaml:"subscription"`
	Clusters     map[string]Cluster `json:"clusters" yaml:"clusters"`
}

func (c ResourceGroup) Cluster(name string) (Cluster, error) {
	value, ok := c.Clusters[name]
	if !ok {
		return Cluster{}, errors.Errorf("cluster not found: %s", name)
	}
	return value, nil
}

func (c ResourceGroup) ClusterNames() []string {
	return lo.Keys(c.Clusters)
}
