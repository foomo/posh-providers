package teleport

import (
	"fmt"
	"os"
	"strings"
)

type (
	Config struct {
		// Directory holding the teleport session state, exported as TELEPORT_HOME
		// when the provider is constructed. Relative to the project root.
		Path string `json:"path" yaml:"path"`
		// Labels every cluster, app and database must carry to be listed. Combined
		// into a single tsh --query expression with &&; omit to list everything.
		Labels map[string]string `json:"labels" yaml:"labels"`
		// Teleport proxy to authenticate against, e.g. "teleport.example.com:443".
		Hostname string `json:"hostname" yaml:"hostname"`
		// Kubernetes-specific settings.
		Kubernetes Kubernetes `json:"kubernetes" yaml:"kubernetes"`
		// Extra arguments passed to `tsh apps login`, keyed by app name. Prepended
		// before the app name itself; apps not listed here still work.
		Apps map[string][]string `json:"apps" yaml:"apps"`
		// Database-specific settings.
		Database Database `json:"database" yaml:"database"`
	}
	Kubernetes struct {
		// Display names for kube clusters, keyed by the real teleport cluster name.
		// Completion offers the alias and it is mapped back before login, so the
		// values must be unique and must not collide with a real cluster name.
		Aliases map[string]string `json:"aliases" yaml:"aliases"`
	}
	Database struct {
		// Database user to log in as. Overridden by the TELEPORT_DATABASE_USER
		// environment variable when it is set.
		User string `json:"user" yaml:"user"`
	}
)

func (c Config) Query() string {
	if len(c.Labels) == 0 {
		return ""
	}

	var ret []string
	for k, v := range c.Labels {
		ret = append(ret, fmt.Sprintf("labels[\"%s\"] == \"%s\"", k, v))
	}

	return strings.Join(ret, " && ")
}

func (c Database) EnvUser() string {
	if value := os.Getenv("TELEPORT_DATABASE_USER"); value != "" {
		return value
	}

	return c.User
}

func (c Kubernetes) Alias(name string) string {
	if c.Aliases == nil {
		return name
	}

	if value, ok := c.Aliases[name]; ok {
		return value
	}

	return name
}

func (c Kubernetes) Name(alias string) string {
	for name, value := range c.Aliases {
		if value == alias {
			return name
		}
	}

	return alias
}
