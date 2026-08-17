package gcx

import (
	"maps"
	"slices"

	"github.com/pkg/errors"
)

type Config map[string]ConfigEnv

// EnvNames returns the names of all environment configurations.
func (c Config) EnvNames() []string {
	return slices.Sorted(maps.Keys(c))
}

func (c Config) Env(name string) (ConfigEnv, error) {
	v, ok := c[name]
	if !ok {
		return ConfigEnv{}, errors.Errorf("missing environment config: %s", name)
	}

	return v, nil
}
