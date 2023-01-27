package webdriverio

import (
	"github.com/foomo/posh-providers/onepassword"
)

type (
	Config struct {
		Secrets map[string]onepassword.Secret `yaml:"secrets"`
		Modes   ConfigModes                   `yaml:"modes"`
		Sites   ConfigSites                   `yaml:"sites"`
		Dirs    []string                      `yaml:"dirs"`
	}
	ConfigModes map[string]ConfigMode
	ConfigMode  struct {
		Port       string `yaml:"port"`
		HostPrefix string `yaml:"hostPrefix"`
	}
	ConfigSites map[string]ConfigEnvs
	ConfigEnvs  map[string]ConfigEnv
	ConfigEnv   struct {
		Domain       string              `yaml:"domain"`
		AuthUsername *onepassword.Secret `yaml:"authUsername"`
		AuthPassword *onepassword.Secret `yaml:"authPassword"`
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
