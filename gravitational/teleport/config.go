package teleport

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
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
		// values must be unique and must not collide with another cluster's real
		// name; a configuration that breaks either rule is rejected when the alias
		// is resolved.
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

// Name maps a display alias back to the real teleport cluster name, returning
// the input unchanged when it is not an alias.
//
// It errors rather than guessing when the configured aliases are ambiguous: the
// lookup used to scan the map's values, so two clusters sharing an alias
// resolved to one of them at random (Go map iteration is unordered), and an
// alias equal to another cluster's real name shadowed it - configuring
// `kubernetes-dev: prod` meant `prod` logged into dev. Both send a login to the
// wrong cluster, so they are rejected instead.
func (c Kubernetes) Name(alias string) (string, error) {
	inverse, err := c.inverseAliases()
	if err != nil {
		return "", err
	}

	if name, ok := inverse[alias]; ok {
		return name, nil
	}

	return alias, nil
}

// inverseAliases builds the alias -> real name mapping, rejecting a
// configuration that cannot resolve unambiguously.
func (c Kubernetes) inverseAliases() (map[string]string, error) {
	ret := make(map[string]string, len(c.Aliases))

	for name, alias := range c.Aliases {
		if existing, ok := ret[alias]; ok {
			first, second := min(existing, name), max(existing, name)

			return nil, errors.Errorf("ambiguous kubernetes alias %q: shared by clusters %q and %q", alias, first, second)
		}

		ret[alias] = name
	}

	for alias, name := range ret {
		if _, ok := c.Aliases[alias]; ok && alias != name {
			return nil, errors.Errorf("kubernetes alias %q collides with the real cluster name %q", alias, alias)
		}
	}

	return ret, nil
}
