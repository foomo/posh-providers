package browser

import (
	"sort"

	"github.com/pkg/errors"
)

type EnvironmentConfig struct {
	// Named SSH proxy from the global proxies config
	Proxy string `json:"proxy" yaml:"proxy"`
	// Default URL to open when no URL argument is given
	URL string `json:"url" yaml:"url"`
	// Open in incognito mode (visual only; session is already isolated via temp profile dir)
	Incognito bool `json:"incognito" yaml:"incognito"`
}

type Config struct {
	// Named environments, each with an optional proxy and default URL
	Environments map[string]EnvironmentConfig `json:"environments" yaml:"environments"`
	// Directory for browser profile dirs (e.g. .posh/browser); created on startup
	ConfigPath string `json:"configPath" yaml:"configPath"`
	// Provider-level default proxy (fallback when env has no proxy set)
	Proxy string `json:"proxy" yaml:"proxy"`
	// Browser binary path; defaults to Chrome on macOS, google-chrome on Linux
	Command string `json:"command" yaml:"command"`
	// Open in incognito mode
	Incognito bool `json:"incognito" yaml:"incognito"`
}

func (c Config) EnvironmentNames() []string {
	names := make([]string, 0, len(c.Environments))
	for name := range c.Environments {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func (c Config) Environment(name string) (EnvironmentConfig, error) {
	env, ok := c.Environments[name]
	if !ok {
		return EnvironmentConfig{}, errors.Errorf("environment not found: %s", name)
	}

	return env, nil
}
