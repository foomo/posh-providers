package webdriverio

import (
	"github.com/foomo/posh-providers/onepassword"
)

type (
	Config struct {
		Dirs         []string                      `json:"dirs" yaml:"dirs"`
		Modes        ConfigModes                   `json:"modes" yaml:"modes"`
		Sites        ConfigSites                   `json:"sites" yaml:"sites"`
		Secrets      map[string]onepassword.Secret `json:"secrets" yaml:"secrets"`
		BrowserStack *onepassword.Secret           `json:"browserStack" yaml:"browserStack"`
	}
	ConfigModes map[string]ConfigMode
	ConfigMode  struct {
		Port       string `json:"port" yaml:"port"`
		HostPrefix string `json:"hostPrefix" yaml:"hostPrefix"`
	}
	ConfigSites map[string]ConfigEnvs
	ConfigEnvs  map[string]ConfigEnv
	ConfigEnv   struct {
		Auth   *onepassword.Secret `json:"auth" yaml:"auth"`
		Domain string              `json:"domain" yaml:"domain"`
	}
)

func (c ConfigSites) Keys() []string {
	ret := make([]string, 0, len(c))
	for s := range c {
		ret = append(ret, s)
	}

	return ret
}

func (c ConfigEnvs) Keys() []string {
	ret := make([]string, 0, len(c))
	for s := range c {
		ret = append(ret, s)
	}

	return ret
}

func (c ConfigModes) Keys() []string {
	ret := make([]string, 0, len(c))
	for s := range c {
		ret = append(ret, s)
	}

	return ret
}
