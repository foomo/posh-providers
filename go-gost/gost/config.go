package gost

import (
	"errors"
	"maps"
	"slices"
)

type Config map[string]string

func (c Config) Names() []string {
	return slices.Sorted(maps.Keys(c))
}

func (c Config) Path(name string) (string, error) {
	v, ok := c[name]
	if !ok {
		return "", errors.New("config key not found")
	}

	return v, nil
}
