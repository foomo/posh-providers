package doctl

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	// Path to the doctl config file holding the API token, relative to the project
	// root. Exported as `DIGITALOCEAN_CONFIG` at shell startup, so every doctl
	// invocation in the session reads this file rather than the user's own; written
	// by `auth init`.
	ConfigPath string `json:"configPath" yaml:"configPath"`
	// Clusters addressable by this command, keyed by the name accepted as the
	// `cluster` argument. The key is a local alias; `name` holds the value passed
	// to doctl.
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
