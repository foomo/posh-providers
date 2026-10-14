package k6

import (
	"sort"

	"github.com/samber/lo"
)

// Config holds the k6 provider's configuration.
type Config struct {
	// Directory holding the `*.k6.js` scenario scripts and their `*.secret.tpl`
	// templates. Created at shell startup if missing, and the root the `scenario`
	// argument is resolved against.
	Path string `json:"path" yaml:"path"`
	// Environments selectable as the first argument, keyed by the name offered in
	// completion. Each entry's keys are upper-cased and passed to k6 as environment
	// variables, so they are visible to the scenario via `__ENV`.
	Envs map[string]Env `json:"envs" yaml:"envs"`
}

// Env returns the environment variables configured for name, or nil if name
// is not a configured environment.
func (c Config) Env(name string) Env {
	return c.Envs[name]
}

// EnvNames returns the configured environment names, sorted alphabetically.
func (c Config) EnvNames() []string {
	ret := lo.Keys(c.Envs)
	sort.Strings(ret)

	return ret
}
