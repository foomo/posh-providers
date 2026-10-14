package gcloud

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// Config is the gcloud provider configuration: the CLOUDSDK_CONFIG path and
// the named accounts and clusters.
type Config struct {
	// Path used as CLOUDSDK_CONFIG, keeping gcloud state per project; created on startup
	ConfigPath string `json:"configPath" yaml:"configPath"`
	// Account configurations, keyed by the name given to `gcloud login`
	Accounts map[string]Account `json:"accounts" yaml:"accounts"`
	// Cluster configurations, keyed by the name given to `gcloud kubeconfig`
	Clusters map[string]Cluster `json:"clusters" yaml:"clusters"`
}

// Cluster returns the cluster registered under name, or an error if none is
// configured for that name.
func (c Config) Cluster(name string) (Cluster, error) {
	value, ok := c.Clusters[name]
	if !ok {
		return Cluster{}, errors.Errorf("given cluster not found: %s", name)
	}

	return value, nil
}

// ClusterNames returns the configured cluster names, in no particular order.
func (c Config) ClusterNames() []string {
	return lo.Keys(c.Clusters)
}

// Account returns the account registered under name, or an error if none is
// configured for that name.
func (c Config) Account(name string) (Account, error) {
	value, ok := c.Accounts[name]
	if !ok {
		return Account{}, errors.Errorf("given account not found: %s", name)
	}

	return value, nil
}

// AccountNames returns the configured account names, in no particular order.
func (c Config) AccountNames() []string {
	return lo.Keys(c.Accounts)
}
