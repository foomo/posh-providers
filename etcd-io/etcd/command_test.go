package etcd_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/etcd-io/etcd"
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
	viper.Set("etcd", map[string]any{
		"configPath": path.Join(dir, "etcd"),
		"clusters": []map[string]any{
			{
				"name":      "my-cluster",
				"podName":   "etcd-0",
				"namespace": "kube-system",
				"paths":     []string{"/registry/config"},
			},
		},
	})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	ec, err := etcd.New(l, kc)
	require.NoError(t, err)

	cmd := etcd.NewCommand(l, ec, kc)

	testutils.AssertDescribed(t, "etcd", cmd)
	testutils.AssertSkilled(t, "etcd", cmd)
}
