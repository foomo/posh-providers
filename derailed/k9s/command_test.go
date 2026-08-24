package k9s_test

import (
	"context"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/derailed/k9s"
	"github.com/foomo/posh-providers/foomo/squadron"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// squadronStub satisfies squadron.Squadron so the fleet and squadron args are
// added to the tree. Only the tree shape matters here; nothing is executed.
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

	// Not parallel: kubectl.New reads its config through viper, which is a singleton.
	viper.Set("kubectl", map[string]any{"configPath": path.Join(t.TempDir(), "kubectl")})
	t.Cleanup(viper.Reset)

	kc, err := kubectl.New(log.NewFmt(), &cache.MemoryCache{})
	require.NoError(t, err)

	// The fleet and squadron args exist only when a squadron is wired, so both
	// wirings are a distinct tree and both ship.
	bare := k9s.NewCommand(log.NewFmt(), kc)

	testutils.AssertDescribed(t, "k9s", bare)
	testutils.AssertSkilled(t, "k9s", bare)

	wired := k9s.NewCommand(log.NewFmt(), kc, k9s.CommandWithSquadron(squadronStub{}))

	testutils.AssertDescribed(t, "k9s (with squadron)", wired)
	testutils.AssertSkilled(t, "k9s (with squadron)", wired)
}
