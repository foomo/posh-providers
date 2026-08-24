package licensefinder_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/pivotal/licensefinder"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	viper.Set("licenseFinder", map[string]any{
		"logPath":       path.Join(dir, "logs"),
		"decisionsPath": path.Join(dir, "licenses.yaml"),
		"sources":       []string{"go.mod"},
	})
	t.Cleanup(viper.Reset)

	cmd, err := licensefinder.NewCommand(log.NewFmt(), &cache.MemoryCache{})
	require.NoError(t, err)

	testutils.AssertDescribed(t, "licensefinder", cmd)
	testutils.AssertSkilled(t, "licensefinder", cmd)
}
