package gost

import (
	"errors"
	"maps"
	"slices"
)

// Config maps a gost process name to the path of the gost configuration file it
// is started with. The keys are the names accepted by `gost start` and
// `gost stop`; each value is passed to gost as `-C <path>` and is resolved
// relative to the project root the posh shell runs in.
type Config map[string]string

// Names returns every configured process name, sorted, which is both the
// suggestion list for the name argument and the set acted on when no name is
// given.
func (c Config) Names() []string {
	return slices.Sorted(maps.Keys(c))
}

// Path returns the gost configuration file path configured for name, or an error
// when name is not a key of the config.
func (c Config) Path(name string) (string, error) {
	v, ok := c[name]
	if !ok {
		return "", errors.New("config key not found")
	}

	return v, nil
}
