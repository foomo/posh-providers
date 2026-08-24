package kubeforward_test

import (
	"log/slog"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/kubernetes/kubeforward"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	viper.Set("kubectl", map[string]any{"configPath": path.Join(t.TempDir(), "kubectl")})
	viper.Set("kubeforward", map[string]any{
		"my-database": map[string]string{
			"cluster":   "my-cluster",
			"namespace": "my-namespace",
			"target":    "service/my-database",
			"port":      "5432:5432",
		},
	})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	cmd, err := kubeforward.NewCommand(l, gokazi.New(slog.New(l.SlogHandler())), kc)
	require.NoError(t, err)

	testutils.AssertDescribed(t, "kubeforward", cmd)
	testutils.AssertSkilled(t, "kubeforward", cmd)
}
