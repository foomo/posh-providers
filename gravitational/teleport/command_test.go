package teleport_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/gravitational/teleport"
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
	// NewTeleport calls os.Setenv("TELEPORT_HOME").
	tmp := t.TempDir()
	viper.Set("kubectl", map[string]any{"configPath": path.Join(tmp, "kubectl")})
	viper.Set("teleport", map[string]any{"path": path.Join(tmp, "teleport")})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()
	c := &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	tp, err := teleport.NewTeleport(l, c)
	require.NoError(t, err)

	cmd := teleport.NewCommand(l, c, tp, kc)

	testutils.AssertDescribed(t, "teleport", cmd)
	testutils.AssertSkilled(t, "teleport", cmd)
}
