package kubectl_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// kubectl.New creates configPath and sets KUBECONFIG, so the key must exist
	// before constructing. Not parallel: viper is a singleton and New mutates the
	// process environment.
	viper.Set("kubectl", map[string]any{"configPath": path.Join(t.TempDir(), "kubectl")})
	t.Cleanup(viper.Reset)

	kc, err := kubectl.New(log.NewFmt(), &cache.MemoryCache{})
	require.NoError(t, err)

	cmd := kubectl.NewCommand(log.NewFmt(), kc)

	testutils.AssertDescribed(t, "kubectl", cmd)
	testutils.AssertSkilled(t, "kubectl", cmd)
}
