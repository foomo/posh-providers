package az

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Subscription struct {
	// Name of the subscription
	Name string `json:"name" yaml:"name"`
	// Vaults available within the subscription
	Vaults map[string]Vault `json:"vaults" yaml:"vaults"`
	// Clusters available within the subscription
	Clusters map[string]Cluster `json:"clusters" yaml:"clusters"`
	// Artifactorys available within the subscription
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

func (c Subscription) Vault(name string) (Vault, error) {
	value, ok := c.Vaults[name]
	if !ok {
		return Vault{}, errors.Errorf("key vault not found: %s", name)
	}

	return value, nil
}

func (c Subscription) VaultNames() []string {
	return lo.Keys(c.Vaults)
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
