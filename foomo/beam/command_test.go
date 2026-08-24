package beam_test

import (
	"log/slog"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh-providers/cloudflare/cloudflared"
	"github.com/foomo/posh-providers/foomo/beam"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// kubectl.New creates configPath and sets KUBECONFIG, so its config key must
	// exist before constructing. Not parallel: viper is a singleton and the
	// constructors mutate the process environment.
	dir := t.TempDir()
	viper.Set("kubectl", map[string]any{"configPath": path.Join(dir, "kubectl")})
	viper.Set("beam", map[string]any{
		"clusters": map[string]any{
			"my-cluster": map[string]any{"port": 12200, "hostname": "my-cluster.example.com"},
		},
		"databases": map[string]any{
			"my-database": map[string]any{"port": 12202, "hostname": "my-database.example.com"},
		},
	})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	op, err := onepassword.New(l, c)
	require.NoError(t, err)

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	cf, err := cloudflared.New(l)
	require.NoError(t, err)

	b, err := beam.New(l, op, gokazi.New(slog.New(l.SlogHandler())))
	require.NoError(t, err)

	cmd, err := beam.NewCommand(l, b, kc, cf)
	require.NoError(t, err)

	testutils.AssertDescribed(t, "beam", cmd)
	testutils.AssertSkilled(t, "beam", cmd)
}
