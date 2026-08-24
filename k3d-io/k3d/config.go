package k3d

import (
	"os"

	"github.com/foomo/posh/pkg/env"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type (
	Config struct {
		// Helm charts installable with the install/uninstall verbs.
		Charts ConfigCharts `json:"charts" yaml:"charts"`
		// Container registry shared by every cluster below. Created on the first
		// `up` and deleted once the last cluster is brought down.
		Registry ConfigRegistry `json:"registry" yaml:"registry"`
		// Clusters that can be created, keyed by the name used as the argument.
		Clusters map[string]ConfigCluster `json:"clusters" yaml:"clusters"`
	}
	ConfigCharts struct {
		// Directory holding one subdirectory per chart, relative to the project
		// root. Its subdirectory names are what the chart argument completes to.
		Path string `json:"path" yaml:"path"`
		// Prefix prepended to the chart name to form the target namespace, so
		// chart "base" with prefix "shared-" installs into "shared-base".
		Prefix string `json:"prefix" yaml:"prefix"`
	}
	ConfigRegistry struct {
		// Registry name passed to k3d, which prefixes it with "k3d-" for the
		// actual container and hostname.
		Name string `json:"name" yaml:"name"`
		// Host port the registry binds to.
		Port string `json:"port" yaml:"port"`
	}
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

func (c Config) Cluster(name string) (ConfigCluster, error) {
	if value, ok := c.Clusters[name]; ok {
		value.Name = name
		return value, nil
	} else {
		return ConfigCluster{}, errors.Errorf("missing cluster config: %s", name)
	}
}

func (c Config) ClusterNames() []string {
	return lo.Keys(c.Clusters)
}

func (c ConfigCluster) AliasName() string {
	if c.Alias == "" {
		return c.Name
	}

	return c.Alias
}
