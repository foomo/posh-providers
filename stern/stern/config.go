package stern

import (
	"sort"

	"github.com/samber/lo"
)

// Config holds the stern provider's configuration.
type Config struct {
	// Named queries selectable by the `query` verb, keyed by the name typed as its
	// first argument. Nested arbitrarily deep: each additional argument descends one
	// level, and every level's `query` arguments are appended in order.
	Queries map[string]Query `json:"queries" yaml:"queries"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// FindQueries walks names as a path into the configured query tree, returning
// every ancestor Query from the first level down to the matched one in order.
// It returns nil if names is empty or the path does not match.
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

// QueryNames returns the sorted query keys one level below the path given by
// names, or the top-level query keys if names is empty or does not match.
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
