package stackit_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh-providers/stackitcloud/stackit"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	viper.Set("kubectl", map[string]any{"configPath": path.Join(dir, "kubectl")})
	viper.Set("stackit", map[string]any{
		"projects": map[string]any{
			"my-project": map[string]any{
				"id": "123456-123456-123456",
				"clusters": map[string]any{
					"dev": map[string]any{"name": "my-project-dev-cluster"},
				},
			},
		},
	})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	sc, err := stackit.New(l, c)
	require.NoError(t, err)

	cmd := stackit.NewCommand(l, c, sc, kc)

	testutils.AssertDescribed(t, "stackit", cmd)
	testutils.AssertSkilled(t, "stackit", cmd)
}
