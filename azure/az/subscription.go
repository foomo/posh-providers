package az

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Subscription struct {
	Name          string                 `json:"name" yaml:"name"`
	Clusters      map[string]Cluster     `json:"clusters" yaml:"clusters"`
	Artifactories map[string]Artifactory `json:"artifactories" yaml:"artifactories"`
}

func (c Subscription) Cluster(name string) (Cluster, error) {
	value, ok := c.Clusters[name]
	if !ok {
		return Cluster{}, errors.Errorf("cluster not found: %s", name)
	}

	return value, nil
}

func (c Subscription) ClusterNames() []string {
	return lo.Keys(c.Clusters)
}

func (c Subscription) Artifactory(name string) (Artifactory, error) {
	value, ok := c.Artifactories[name]
	if !ok {
		return Artifactory{}, errors.Errorf("artifactory not found: %s", name)
	}

	return value, nil
}

func (c Subscription) ArtifactoryNames() []string {
	return lo.Keys(c.Artifactories)
}
