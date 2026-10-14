package chrome

import (
	"sort"

	"github.com/pkg/errors"
)

// Config is the chrome provider configuration: the profile data directory,
// an optional browser binary override, a default incognito setting, and the
// named profiles.
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

// ProfileNames returns the configured profile names, sorted alphabetically.
func (c Config) ProfileNames() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// Profile returns the profile registered under name, or an error if none is
// configured for that name.
func (c Config) Profile(name string) (Profile, error) {
	env, ok := c.Profiles[name]
	if !ok {
		return Profile{}, errors.Errorf("profile not found: %s", name)
	}

	return env, nil
}
