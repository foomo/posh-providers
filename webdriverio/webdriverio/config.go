package webdriverio

import (
	"github.com/foomo/posh-providers/onepassword"
)

type (
	// Config holds the modes, sites and secrets the wdio command runs against.
	Config struct {
		// Unused. Test directories are discovered by walking for wdio.conf.ts,
		// never read from here.
		Dirs []string `json:"dirs" yaml:"dirs"`
		// Run modes, keyed by name and suggested as the first argument. The key
		// "browserstack" additionally selects the BrowserStack backend.
		Modes ConfigModes `json:"modes" yaml:"modes"`
		// Sites to test, keyed by name and suggested as the second argument.
		Sites ConfigSites `json:"sites" yaml:"sites"`
		// Unused. Declared but read by nothing; only Sites[..].Auth and
		// BrowserStack resolve secrets.
		Secrets map[string]onepassword.Secret `json:"secrets" yaml:"secrets"`
		// 1Password item holding the BrowserStack credentials, read from its
		// "username" and "password" fields. Required when the "browserstack"
		// mode is used without --ci; the command panics if it is unset.
		BrowserStack *onepassword.Secret `json:"browserStack" yaml:"browserStack"`
	}
	// ConfigModes maps a run mode name to its settings.
	ConfigModes map[string]ConfigMode
	// ConfigMode holds the host prefix and port a run mode applies to a site's
	// domain.
	ConfigMode struct {
		// Port appended to the site domain, e.g. "8443". Omit for the default port.
		Port string `json:"port" yaml:"port"`
		// Subdomain prepended to the site domain, e.g. "local". Omit for none.
		HostPrefix string `json:"hostPrefix" yaml:"hostPrefix"`
	}
	// ConfigSites maps a site name to its configured environments.
	ConfigSites map[string]ConfigEnvs
	// ConfigEnvs maps an environment name to its settings.
	ConfigEnvs map[string]ConfigEnv
	// ConfigEnv holds the domain and optional basic auth secret for one site
	// environment.
	ConfigEnv struct {
		// 1Password item holding basic auth credentials for this environment,
		// read from its "username" and "password" fields. Omit for no basic auth.
		Auth *onepassword.Secret `json:"auth" yaml:"auth"`
		// Base domain for this environment, e.g. "staging.example.com". The mode's
		// host prefix and port are applied on top of it.
		Domain string `json:"domain" yaml:"domain"`
	}
)

// Keys returns the configured site names.
func (c ConfigSites) Keys() []string {
	ret := make([]string, 0, len(c))
	for s := range c {
		ret = append(ret, s)
	}

	return ret
}

// Keys returns the configured environment names.
func (c ConfigEnvs) Keys() []string {
	ret := make([]string, 0, len(c))
	for s := range c {
		ret = append(ret, s)
	}

	return ret
}

// Keys returns the configured mode names.
func (c ConfigModes) Keys() []string {
	ret := make([]string, 0, len(c))
	for s := range c {
		ret = append(ret, s)
	}

	return ret
}
