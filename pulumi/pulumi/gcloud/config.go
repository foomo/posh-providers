package pulumi

import (
	"github.com/pkg/errors"
)

type Config struct {
	Path       string             `json:"path" yaml:"path"`
	ConfigPath string             `json:"configPath" yaml:"configPath"`
	Backends   map[string]Backend `json:"backends" yaml:"backends"`
}

func (p Config) Backend(name string) (Backend, error) {
	value, ok := p.Backends[name]
	if !ok {
		return Backend{}, errors.Errorf("backend not found: %s", name)
	}

	return value, nil
}
