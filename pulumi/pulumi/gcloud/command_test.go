package pulumi_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/google/gcloud"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/pkg/testutils"
	pulumi "github.com/foomo/posh-providers/pulumi/pulumi/gcloud"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	viper.Set("pulumi", map[string]any{
		"path":       path.Join(dir, "pulumi"),
		"configPath": path.Join(dir, "config", "pulumi"),
	})
	viper.Set("gcloud", map[string]any{"configPath": path.Join(dir, "gcloud")})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	gc, err := gcloud.New(l, c)
	require.NoError(t, err)

	op, err := onepassword.New(l, c)
	require.NoError(t, err)

	cmd, err := pulumi.NewCommand(l, gc, op, c)
	require.NoError(t, err)

	testutils.AssertDescribed(t, "pulumi", cmd)
	testutils.AssertSkilled(t, "pulumi", cmd)
}
