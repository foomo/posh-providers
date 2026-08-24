package az_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/azure/az"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// az.New creates the parent of configPath and sets AZURE_CONFIG_DIR;
	// kubectl.New creates configPath and sets KUBECONFIG - so both provider
	// config keys must exist before constructing. Not parallel: both mutate the
	// process environment and viper is a singleton.
	dir := t.TempDir()
	viper.Set("az", map[string]any{"configPath": path.Join(dir, "azure", "config")})
	viper.Set("kubectl", map[string]any{"configPath": path.Join(dir, "kubectl")})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()
	c := &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	a, err := az.New(l, c)
	require.NoError(t, err)

	cmd := az.NewCommand(l, a, kc)

	testutils.AssertDescribed(t, "az", cmd)
	testutils.AssertSkilled(t, "az", cmd)
}
