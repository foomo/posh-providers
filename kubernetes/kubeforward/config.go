package kubeforward

import (
	"sort"

	"github.com/samber/lo"
)

// Config holds the configured port forwards, keyed by name.
type Config map[string]PortForward

// Names returns the configured port forward names, sorted alphabetically.
func (c Config) Names() []string {
	ret := lo.Keys(c)
	sort.Strings(ret)

	return ret
}
