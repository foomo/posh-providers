package stern_test

import (
	"context"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/foomo/squadron"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh-providers/stern/stern"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// squadronStub satisfies squadron.Squadron, which NewCommand requires. Only the
// tree shape matters here; nothing is executed.
type squadronStub struct{}

func (squadronStub) Cluster(string) (squadron.Cluster, bool) { return squadron.Cluster{}, false }
func (squadronStub) Exists(string) bool                      { return false }
func (squadronStub) List() ([]string, error)                 { return nil, nil }
func (squadronStub) UnitDirs(string) []string                { return nil }
func (squadronStub) GetFiles(string, string, string, bool) []string {
	return nil
}

func (squadronStub) UnitExists(context.Context, string, string, string, string, bool) bool {
	return false
}

func (squadronStub) ListUnits(context.Context, string, string, string, bool) ([]string, error) {
	return nil, nil
}

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Not parallel: kubectl.New and the provider both read config through viper.
	viper.Set("kubectl", map[string]any{"configPath": path.Join(t.TempDir(), "kubectl")})
	viper.Set("stern", map[string]any{
		"queries": map[string]any{
			"all": map[string]any{
				"query": []string{".*", "--all-namespaces"},
				"queries": map[string]any{
					"errors": map[string]any{"query": []string{"--include", "error"}},
				},
			},
		},
	})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()

	kc, err := kubectl.New(l, &cache.MemoryCache{})
	require.NoError(t, err)

	cmd, err := stern.NewCommand(l, kc, squadronStub{})
	require.NoError(t, err)

	testutils.AssertDescribed(t, "stern", cmd)
	testutils.AssertSkilled(t, "stern", cmd)
}
