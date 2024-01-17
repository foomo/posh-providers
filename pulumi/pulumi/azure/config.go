package pulumi

import (
	"github.com/foomo/posh-providers/onepassword"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	Path       string             `json:"path" yaml:"path"`
	ConfigPath string             `json:"configPath" yaml:"configPath"`
	Passphrase onepassword.Secret `json:"passphrase" yaml:"passphrase"`
	Backends   map[string]Backend `json:"backends" yaml:"backends"`
}

func (p Config) Backend(name string) (Backend, error) {
	value, ok := p.Backends[name]
	if !ok {
		return Backend{}, errors.Errorf("backend not found: %s", name)
	}
	return value, nil
}

func (p Config) Azure() []string {
	return lo.Keys(p.Backends)
}
