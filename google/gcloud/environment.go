package gcloud

import (
	"github.com/pkg/errors"
)

type Environment struct {
	Name     string    `json:"name" yaml:"name"`
	Project  string    `json:"project" yaml:"project"`
	Clusters []Cluster `json:"clusters" yaml:"clusters"`
}

func (e Environment) Cluster(name string) (Cluster, error) {
	for _, cluster := range e.Clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}
	return Cluster{}, errors.Errorf("given cluster not found: %s", name)
}

func (e Environment) ClusterNames() []string {
	ret := make([]string, len(e.Clusters))
	for i, cluster := range e.Clusters {
		ret[i] = cluster.Name
	}
	return ret
}
