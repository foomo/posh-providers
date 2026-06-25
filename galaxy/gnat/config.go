package gnat

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	// Theme is the gnat color theme, passed to the CLI via -theme.
	Theme string `json:"theme,omitempty" yaml:"theme,omitempty"`
	// URLs maps a name to a NATS server URL; names autocomplete in the shell.
	URLs map[string]string `json:"urls" yaml:"urls"`
}

func (c Config) URLNames() []string {
	return lo.Keys(c.URLs)
}

func (c Config) URL(name string) (string, error) {
	value, ok := c.URLs[name]
	if !ok {
		return "", errors.Errorf("given url not found: %s", name)
	}

	return value, nil
}
