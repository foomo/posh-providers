package gcloud

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	// Path used as CLOUDSDK_CONFIG, keeping gcloud state per project; created on startup
	ConfigPath string `json:"configPath" yaml:"configPath"`
	// Account configurations, keyed by the name given to `gcloud login`
	Accounts map[string]Account `json:"accounts" yaml:"accounts"`
	// Cluster configurations, keyed by the name given to `gcloud kubeconfig`
	Clusters map[string]Cluster `json:"clusters" yaml:"clusters"`
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

func (c Config) Account(name string) (Account, error) {
	value, ok := c.Accounts[name]
	if !ok {
		return Account{}, errors.Errorf("given account not found: %s", name)
	}

	return value, nil
}

func (c Config) AccountNames() []string {
	return lo.Keys(c.Accounts)
}
