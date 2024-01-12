package k3d

import (
	"os"

	"github.com/foomo/posh/pkg/env"
	"github.com/pkg/errors"
)

type (
	Config struct {
		Charts   ConfigCharts    `json:"charts" yaml:"charts"`
		Registry ConfigRegistry  `json:"registry" yaml:"registry"`
		Clusters []ConfigCluster `json:"clusters" yaml:"clusters"`
	}
	ConfigCharts struct {
		Path   string `json:"path" yaml:"path"`
		Prefix string `json:"prefix" yaml:"prefix"`
	}
	ConfigRegistry struct {
		Name string `json:"name" yaml:"name"`
		Port string `json:"port" yaml:"port"`
	}
	ConfigCluster struct {
		// Cluster name
		Name string `json:"name" yaml:"name"`
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

func (c Config) Cluster(name string) (ConfigCluster, error) {
	for _, v := range c.Clusters {
		if v.Name == name {
			return v, nil
		}
	}
	return ConfigCluster{}, errors.Errorf("missing cluster config: %s", name)
}

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

func (c Config) ClusterNames() []string {
	ret := make([]string, len(c.Clusters))
	for i, v := range c.Clusters {
		ret[i] = v.Name
	}
	return ret
}

func (c ConfigCluster) AliasName() string {
	if c.Alias == "" {
		return c.Name
	}
	return c.Alias
}
