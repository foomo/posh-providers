package stern

import (
	"sort"

	"github.com/samber/lo"
)

// Query is one level of a nested, named stern argument tree.
type Query struct {
	// Arguments passed verbatim to stern when this level is selected, e.g. a pod
	// regex plus flags. Concatenated with every ancestor level's arguments, so a
	// nested entry refines its parent rather than replacing it.
	Query []string `json:"query" yaml:"query"`
	// Further queries nested under this one, selected by appending their name as
	// another argument.
	Queries map[string]Query `json:"queries" yaml:"queries"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// FindQueries walks names as a path into this Query's nested queries,
// returning every ancestor Query from the first level down to the matched one
// in order. It returns nil if names is empty or the path does not match.
func (q Query) FindQueries(names ...string) []Query {
	if len(names) == 0 {
		return nil
	}

	query, ok := q.Queries[names[0]]
	if !ok {
		return nil
	}

	ret := []Query{query}
	if queries := query.FindQueries(names[1:]...); queries != nil {
		ret = append(ret, queries...)
	}

	return ret
}

// QueryNames returns the sorted nested query keys one level below the path
// given by names, or this Query's own nested keys if names is empty or does
// not match.
func (q Query) QueryNames(names ...string) []string {
	if len(names) == 0 {
		ret := lo.Keys(q.Queries)
		sort.Strings(ret)

		return ret
	}

	query, ok := q.Queries[names[0]]
	if !ok {
		ret := lo.Keys(q.Queries)
		sort.Strings(ret)

		return ret
	}

	return query.QueryNames(names[1:]...)
}
