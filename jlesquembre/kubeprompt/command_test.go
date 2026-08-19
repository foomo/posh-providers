package kubeprompt_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/jlesquembre/kubeprompt"
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
	// before constructing. Not parallel: viper is a singleton and kubectl.New
	// mutates the process environment.
	viper.Set("kubectl", map[string]any{"configPath": path.Join(t.TempDir(), "kubectl")})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()

	kc, err := kubectl.New(l, &cache.MemoryCache{})
	require.NoError(t, err)

	cmd := kubeprompt.NewCommand(l, kc)

	testutils.AssertDescribed(t, "kubeprompt", cmd)
	testutils.AssertSkilled(t, "kubeprompt", cmd)
}
