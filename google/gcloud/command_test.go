package gcloud_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/google/gcloud"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// GCloud.New creates ConfigPath and sets CLOUDSDK_CONFIG, so both provider
	// config keys must exist before constructing. Not parallel: New mutates the
	// process environment. Verified working.
	dir := t.TempDir()
	viper.Set("gcloud", map[string]any{"configPath": path.Join(dir, "gcloud")})
	viper.Set("kubectl", map[string]any{"configPath": path.Join(dir, "kubectl")})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()
	c := &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	gc, err := gcloud.New(l, c)
	require.NoError(t, err)

	cmd := gcloud.NewCommand(l, gc, kc)

	testutils.AssertDescribed(t, "gcloud", cmd)
	testutils.AssertSkilled(t, "gcloud", cmd)
}
