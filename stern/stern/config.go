package stern

import (
	"sort"

	"github.com/samber/lo"
)

type Config struct {
	Queries map[string]Query `json:"queries" yaml:"queries"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c Config) FindQueries(names ...string) []Query {
	if len(names) == 0 {
		return nil
	}

	query, ok := c.Queries[names[0]]
	if !ok {
		return nil
	}

	ret := []Query{query}
	if queries := query.FindQueries(names[1:]...); queries != nil {
		ret = append(ret, queries...)
	}

	return ret
}

func (c Config) QueryNames(names ...string) []string {
	if len(names) == 0 {
		ret := lo.Keys(c.Queries)
		sort.Strings(ret)

		return ret
	}

	query, ok := c.Queries[names[0]]
	if !ok {
		ret := lo.Keys(c.Queries)
		sort.Strings(ret)

		return ret
	}

	return query.QueryNames(names[1:]...)
}
