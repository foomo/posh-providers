package k6_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/grafana/k6"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// New() creates cfg.Path, so keep it out of the checkout.
	dir := t.TempDir()
	viper.Set("k6", map[string]any{
		"path": path.Join(dir, "k6"),
		"envs": map[string]any{
			"dev": map[string]any{"URL": "https://example.invalid"},
		},
	})
	viper.Set("onePassword", map[string]any{"account": "example"})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()

	op, err := onepassword.New(l, &cache.MemoryCache{})
	require.NoError(t, err)

	cmd, err := k6.NewCommand(l, op)
	require.NoError(t, err)

	testutils.AssertDescribed(t, "k6", cmd)
	testutils.AssertSkilled(t, "k6", cmd)
}
