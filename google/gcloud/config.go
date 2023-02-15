package gcloud

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	ConfigPath string             `json:"configPath" yaml:"configPath"`
	Accounts   map[string]Account `json:"accounts" yaml:"accounts"`
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
