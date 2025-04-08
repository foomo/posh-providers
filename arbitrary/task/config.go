package task

import (
	"sort"
)

type Config map[string]Task

func (c Config) Names() []string {
	var ret []string
	for k, v := range c {
		if !v.Hidden {
			ret = append(ret, k)
		}
	}
	sort.Strings(ret)
	return ret
}
