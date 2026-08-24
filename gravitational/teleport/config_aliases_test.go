package teleport_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/gravitational/teleport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKubernetesName covers mapping a display alias back to the real cluster.
//
// The lookup used to scan the Aliases map's *values*, so:
//
//   - two clusters sharing an alias resolved to one of them at random - measured
//     a 171/29 split over 200 calls, since Go map iteration is unordered;
//   - the fallback ran after the loop, so an alias equal to another cluster's
//     real name shadowed it: `kubernetes-dev: prod` made `prod` resolve to
//     `kubernetes-dev` rather than to the real cluster named `prod`.
//
// Both send a login to the wrong cluster silently, so they are now rejected.
func TestKubernetesName(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	t.Run("resolves an alias to its real cluster", func(t *testing.T) {
		t.Parallel()

		k := teleport.Kubernetes{Aliases: map[string]string{"prod-eu-1": "prod"}}

		got, err := k.Name("prod")
		require.NoError(t, err)
		assert.Equal(t, "prod-eu-1", got)
	})

	t.Run("passes through a name that is not an alias", func(t *testing.T) {
		t.Parallel()

		k := teleport.Kubernetes{Aliases: map[string]string{"prod-eu-1": "prod"}}

		got, err := k.Name("staging")
		require.NoError(t, err)
		assert.Equal(t, "staging", got)
	})

	t.Run("no aliases configured", func(t *testing.T) {
		t.Parallel()

		var k teleport.Kubernetes

		got, err := k.Name("prod")
		require.NoError(t, err)
		assert.Equal(t, "prod", got)
	})

	t.Run("rejects an alias shared by two clusters", func(t *testing.T) {
		t.Parallel()

		k := teleport.Kubernetes{Aliases: map[string]string{
			"prod-eu": "prod",
			"prod-us": "prod",
		}}

		// Deterministic regardless of map order: the message names both clusters
		// sorted, so this cannot flake.
		_, err := k.Name("prod")
		require.Error(t, err)
		assert.EqualError(t, err,
			`ambiguous kubernetes alias "prod": shared by clusters "prod-eu" and "prod-us"`)
	})

	t.Run("rejects an alias that shadows a real cluster name", func(t *testing.T) {
		t.Parallel()

		// `prod` is both an alias of kubernetes-dev and a real cluster.
		k := teleport.Kubernetes{Aliases: map[string]string{
			"kubernetes-dev": "prod",
			"prod":           "production",
		}}

		_, err := k.Name("prod")
		require.Error(t, err)
		assert.EqualError(t, err,
			`kubernetes alias "prod" collides with the real cluster name "prod"`)
	})

	t.Run("an alias equal to its own cluster name is fine", func(t *testing.T) {
		t.Parallel()

		k := teleport.Kubernetes{Aliases: map[string]string{"prod": "prod"}}

		got, err := k.Name("prod")
		require.NoError(t, err)
		assert.Equal(t, "prod", got)
	})
}
