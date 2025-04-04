package k6

import (
	"os"
	"sort"
	"strings"

	"github.com/foomo/posh/pkg/env"
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

func (c Config) Scenarios() ([]string, error) {
	entries, err := os.ReadDir(env.Path(c.Path))
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".js") {
			ret = append(ret, e.Name())
		}
	}
	return ret, nil
}
