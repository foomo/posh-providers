package stern

import (
	"sort"

	"github.com/samber/lo"
)

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
