package k3d

import (
	"os"

	"github.com/foomo/posh/pkg/env"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type (
	// Config holds the k3d provider's settings.
	Config struct {
		// Helm charts installable with the install/uninstall verbs.
		Charts ConfigCharts `json:"charts" yaml:"charts"`
		// Container registry shared by every cluster below. Created on the first
		// `up` and deleted once the last cluster is brought down.
		Registry ConfigRegistry `json:"registry" yaml:"registry"`
		// Clusters that can be created, keyed by the name used as the argument.
		Clusters map[string]ConfigCluster `json:"clusters" yaml:"clusters"`
	}
	// ConfigCharts configures where predefined helm charts live and how their
	// target namespace is named.
	ConfigCharts struct {
		// Directory holding one subdirectory per chart, relative to the project
		// root. Its subdirectory names are what the chart argument completes to.
		Path string `json:"path" yaml:"path"`
		// Prefix prepended to the chart name to form the target namespace, so
		// chart "base" with prefix "shared-" installs into "shared-base".
		Prefix string `json:"prefix" yaml:"prefix"`
	}
	// ConfigRegistry configures the shared container registry used by every
	// cluster.
	ConfigRegistry struct {
		// Registry name passed to k3d, which prefixes it with "k3d-" for the
		// actual container and hostname.
		Name string `json:"name" yaml:"name"`
		// Host port the registry binds to.
		Port string `json:"port" yaml:"port"`
	}
	// ConfigCluster configures a single creatable cluster.
	ConfigCluster struct {
		// Cluster name
		Name string `json:"-" yaml:"-"`
		// K3d cluster name
		Alias string `json:"alias" yaml:"alias"`
		// Docker image to use
		Image string `json:"image" yaml:"image"`
		// Port to bind to
		Port string `json:"port" yaml:"port"`
		// EnableTraefikRouter allows to create the cluster with the default traefik router
		EnableTraefikRouter bool `json:"enableTraefikRouter" yaml:"enableTraefikRouter"`
		// Additional arguments
		Args []string `json:"args" yaml:"args"`
	}
)

// Names returns the subdirectory names under the charts path, i.e. the
// installable chart names.
func (c ConfigCharts) Names() ([]string, error) {
	files, err := os.ReadDir(env.Path(c.Path))
	if err != nil {
		return nil, err
	}

	var ret []string

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		ret = append(ret, file.Name())
	}

	return ret, nil
}

// Cluster returns the ConfigCluster registered under name, with its Name
// field set, or an error if no such cluster is configured.
func (c Config) Cluster(name string) (ConfigCluster, error) {
	if value, ok := c.Clusters[name]; ok {
		value.Name = name
		return value, nil
	} else {
		return ConfigCluster{}, errors.Errorf("missing cluster config: %s", name)
	}
}

// ClusterNames returns the configured cluster names.
func (c Config) ClusterNames() []string {
	return lo.Keys(c.Clusters)
}

// AliasName returns the cluster's Alias if set, otherwise its Name. This is
// the name k3d itself knows the cluster by.
func (c ConfigCluster) AliasName() string {
	if c.Alias == "" {
		return c.Name
	}

	return c.Alias
}
