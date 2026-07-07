package tempo

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	// Theme is the tempo color theme, passed to the CLI via --theme.
	Theme string `json:"theme,omitempty" yaml:"theme,omitempty"`
	// ConfigDir specifies the $XDG_CONFIG_HOME where configuration files are stored.
	ConfigDir string `json:"configDir" yaml:"configDir"`
	// Profiles maps a name to a Temporal connection profile; names autocomplete in the shell.
	Profiles map[string]Profile `json:"profiles" yaml:"profiles"`
}

func (c Config) ProfileNames() []string {
	return lo.Keys(c.Profiles)
}

func (c Config) Profile(name string) (Profile, error) {
	value, ok := c.Profiles[name]
	if !ok {
		return Profile{}, errors.Errorf("given profile not found: %s", name)
	}

	return value, nil
}
