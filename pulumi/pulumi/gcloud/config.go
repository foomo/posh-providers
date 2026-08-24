package pulumi

import (
	"github.com/pkg/errors"
)

type Config struct {
	// Path to the directory holding one sub directory per environment, each
	// containing one sub directory per pulumi project. Environment and project
	// completion read this tree from disk, so an environment that is configured
	// under Backends but missing here cannot be selected.
	Path string `json:"path" yaml:"path"`
	// ConfigPath is exported as PULUMI_HOME, relocating pulumi's own state
	// (credentials, plugins, templates) into the project. Resolved through
	// env.Path, so a relative value is taken from the project root.
	ConfigPath string `json:"configPath" yaml:"configPath"`
	// Backends maps an environment name to the Google Cloud Storage bucket
	// holding its pulumi state. The key is the <env> argument, so it must match
	// a directory name below Path.
	Backends map[string]Backend `json:"backends" yaml:"backends"`
}

func (p Config) Backend(name string) (Backend, error) {
	value, ok := p.Backends[name]
	if !ok {
		return Backend{}, errors.Errorf("backend not found: %s", name)
	}

	return value, nil
}
