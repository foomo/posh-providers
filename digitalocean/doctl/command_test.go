package doctl_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/digitalocean/doctl"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	viper.Set("kubectl", map[string]any{"configPath": path.Join(dir, "kubectl")})
	viper.Set("doctl", map[string]any{
		"configPath": path.Join(dir, "doctl", "doctl.yaml"),
		"clusters": map[string]any{
			"prod": map[string]any{"name": "00000000-0000-0000-0000-000000000000"},
		},
	})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	dc, err := doctl.New(l, c)
	require.NoError(t, err)

	cmd := doctl.NewCommand(l, c, dc, kc)

	testutils.AssertDescribed(t, "doctl", cmd)
	testutils.AssertSkilled(t, "doctl", cmd)
}
