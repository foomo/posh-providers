package stackit

import (
	"sort"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Project struct {
	ID       string             `json:"id" yaml:"id"`
	Clusters map[string]Cluster `json:"clusters" yaml:"clusters"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c Project) Cluster(name string) (Cluster, error) {
	value, ok := c.Clusters[name]
	if !ok {
		return Cluster{}, errors.Errorf("given cluster not found: %s", name)
	}

	return value, nil
}

func (c Project) ClusterNames() []string {
	ret := lo.Keys(c.Clusters)
	sort.Strings(ret)

	return ret
}
