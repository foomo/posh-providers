package az

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// Subscription holds an Azure subscription's name and the vaults, clusters
// and artifactories configured within it.
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

// Cluster returns the named cluster configuration. It returns an error if the
// name is not configured.
func (c Subscription) Cluster(name string) (Cluster, error) {
	value, ok := c.Clusters[name]
	if !ok {
		return Cluster{}, errors.Errorf("cluster not found: %s", name)
	}

	return value, nil
}

// ClusterNames returns the configured cluster names.
func (c Subscription) ClusterNames() []string {
	return lo.Keys(c.Clusters)
}

// Vault returns the named key vault configuration. It returns an error if the
// name is not configured.
func (c Subscription) Vault(name string) (Vault, error) {
	value, ok := c.Vaults[name]
	if !ok {
		return Vault{}, errors.Errorf("key vault not found: %s", name)
	}

	return value, nil
}

// VaultNames returns the configured key vault names.
func (c Subscription) VaultNames() []string {
	return lo.Keys(c.Vaults)
}

// Artifactory returns the named artifactory configuration. It returns an
// error if the name is not configured.
func (c Subscription) Artifactory(name string) (Artifactory, error) {
	value, ok := c.Artifactories[name]
	if !ok {
		return Artifactory{}, errors.Errorf("artifactory not found: %s", name)
	}

	return value, nil
}

// ArtifactoryNames returns the configured artifactory names.
func (c Subscription) ArtifactoryNames() []string {
	return lo.Keys(c.Artifactories)
}
