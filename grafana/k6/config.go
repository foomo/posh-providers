package k6

import (
	"sort"

	"github.com/samber/lo"
)

type Config struct {
	Path string         `json:"path" yaml:"path"`
	Envs map[string]Env `json:"envs" yaml:"envs"`
}

func (c Config) Env(name string) Env {
	return c.Envs[name]
}

func (c Config) EnvNames() []string {
	ret := lo.Keys(c.Envs)
	sort.Strings(ret)

	return ret
}
