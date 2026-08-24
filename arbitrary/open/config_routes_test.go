package open_test

import (
	"maps"
	"slices"
	"strings"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/arbitrary/open"
	"github.com/stretchr/testify/assert"
)

// TestRoutesForPath covers descending into nested routes.
//
// The loop used to `break` inside the `if`, so it stopped after the first match
// instead of continuing the descent: RoutesForPath returned the level-1 children
// whatever the depth, and RouteForPath resolved to a zero ConfigRoute from the
// third level on. Validate rejects the empty Path, so nothing wrong was ever
// opened - the deeper routes were simply unreachable and completion there was
// empty.
func TestRoutesForPath(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	router := open.ConfigRouter{
		URL: "https://example.com",
		Routes: map[string]open.ConfigRoute{
			"a": {Path: "/a", Routes: map[string]open.ConfigRoute{
				"b": {Path: "/a/b", Routes: map[string]open.ConfigRoute{
					"c": {Path: "/a/b/c", Routes: map[string]open.ConfigRoute{
						"d": {Path: "/a/b/c/d"},
					}},
				}},
			}},
		},
	}

	tests := []struct {
		paths    []string
		wantPath string
		children []string
	}{
		{[]string{"a"}, "/a", []string{"b"}},
		{[]string{"a", "b"}, "/a/b", []string{"c"}},
		{[]string{"a", "b", "c"}, "/a/b/c", []string{"d"}},
		{[]string{"a", "b", "c", "d"}, "/a/b/c/d", nil},
		// an unknown segment stops the descent and yields no route, which
		// Validate turns into `invalid [route] argument`
		{[]string{"a", "zz"}, "", []string{"b"}},
	}

	for _, tt := range tests {
		t.Run(joinPaths(tt.paths), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantPath, router.RouteForPath(tt.paths).Path)

			got := slices.Sorted(maps.Keys(router.RoutesForPath(tt.paths)))
			if tt.children == nil {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.children, got)
			}
		})
	}
}

func joinPaths(paths []string) string {
	return strings.Join(paths, "/")
}
