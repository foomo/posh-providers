package gcloud

import (
	"github.com/pkg/errors"
)

type Config struct {
	ConfigPath      string        `json:"configPath" yaml:"configPath"`
	AccessTokenPath string        `json:"accessTokenPath" yaml:"accessTokenPath"`
	Environments    []Environment `json:"environments" yaml:"environments"`
}

func (c Config) Environment(name string) (Environment, error) {
	for _, environment := range c.Environments {
		if environment.Name == name {
			return environment, nil
		}
	}
	return Environment{}, errors.Errorf("given environment not found: %s", name)
}

func (c Config) EnvironmentNames() []string {
	ret := make([]string, len(c.Environments))
	for i, environment := range c.Environments {
		ret[i] = environment.Name
	}
	return ret
}

func (c Config) AllEnvironmentsClusterNames() []string {
	var ret []string
	for _, environment := range c.Environments {
		ret = append(ret, environment.ClusterNames()...)
	}
	return ret
}
