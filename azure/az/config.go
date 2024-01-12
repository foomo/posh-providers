package az

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	ConfigPath     string                   `json:"configPath" yaml:"configPath"`
	ResourceGroups map[string]ResourceGroup `json:"resourceGroups" yaml:"resourceGroups"`
}

func (c Config) ResourceGroup(name string) (ResourceGroup, error) {
	value, ok := c.ResourceGroups[name]
	if !ok {
		return ResourceGroup{}, errors.Errorf("resource group not found: %s", name)
	}
	return value, nil
}

func (c Config) ResourceGroupNames() []string {
	return lo.Keys(c.ResourceGroups)
}
