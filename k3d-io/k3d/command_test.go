package k3d_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/k3d-io/k3d"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Not parallel: both constructors read viper, which is a singleton, and
	// k3d.New calls os.Setenv("K3D_FIX_DNS").
	viper.Set("kubectl", map[string]any{"configPath": path.Join(t.TempDir(), "kubectl")})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()
	c := &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	kd, err := k3d.New(l)
	require.NoError(t, err)

	cmd, err := k3d.NewCommand(l, kd, c, kc)
	require.NoError(t, err)

	testutils.AssertDescribed(t, "k3d", cmd)
	testutils.AssertSkilled(t, "k3d", cmd)
}
