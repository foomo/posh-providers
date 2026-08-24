package stackit

import (
	"sort"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	// STACKIT projects addressable by this command, keyed by the name accepted as
	// the `project` argument. The key is a local alias and need not match the
	// project's name in STACKIT.
	Projects map[string]Project `json:"projects" yaml:"projects"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c Config) ProjectNames() []string {
	ret := lo.Keys(c.Projects)
	sort.Strings(ret)

	return ret
}

func (c Config) Project(name string) (Project, error) {
	value, ok := c.Projects[name]
	if !ok {
		return Project{}, errors.Errorf("given project not found: %s", name)
	}

	return value, nil
}
