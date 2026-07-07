package chrome

import (
	"sort"

	"github.com/pkg/errors"
)

type Config struct {
	// Directory for browser profile dirs (e.g. .posh/browser); created on startup
	Path string `json:"path" yaml:"path"`
	// Browser binary path; defaults to Chrome on macOS, google-chrome on Linux
	Command string `json:"command" yaml:"command"`
	// Open in incognito mode
	Incognito bool `json:"incognito" yaml:"incognito"`
	// Named environments, each with an optional proxy and default URL
	Profiles map[string]Profile `json:"profiles" yaml:"profiles"`
}

func (c Config) ProfileNames() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func (c Config) Profile(name string) (Profile, error) {
	env, ok := c.Profiles[name]
	if !ok {
		return Profile{}, errors.Errorf("profile not found: %s", name)
	}

	return env, nil
}
