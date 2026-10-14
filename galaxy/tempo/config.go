package tempo

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// Config is the tempo provider configuration: the color theme, an optional
// XDG config directory, and the named Temporal connection profiles.
type Config struct {
	// Theme is the tempo color theme, passed to the CLI via --theme.
	Theme string `json:"theme,omitempty" yaml:"theme,omitempty"`
	// ConfigDir specifies the $XDG_CONFIG_HOME where configuration files are stored.
	ConfigDir string `json:"configDir" yaml:"configDir"`
	// Profiles maps a name to a Temporal connection profile; names autocomplete in the shell.
	Profiles map[string]Profile `json:"profiles" yaml:"profiles"`
}

// ProfileNames returns the configured profile names, in no particular order.
func (c Config) ProfileNames() []string {
	return lo.Keys(c.Profiles)
}

// Profile returns the profile registered under name, or an error if none is
// configured for that name.
func (c Config) Profile(name string) (Profile, error) {
	value, ok := c.Profiles[name]
	if !ok {
		return Profile{}, errors.Errorf("given profile not found: %s", name)
	}

	return value, nil
}
