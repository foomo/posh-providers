package kubeforward

import (
	"sort"

	"github.com/samber/lo"
)

type Config map[string]PortForward

func (c Config) Names() []string {
	ret := lo.Keys(c)
	sort.Strings(ret)
	return ret
}
