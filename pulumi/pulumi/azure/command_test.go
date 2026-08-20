package pulumi_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/azure/az"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/pkg/testutils"
	pulumi "github.com/foomo/posh-providers/pulumi/pulumi/azure"
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
	viper.Set("az", map[string]any{"configPath": path.Join(dir, "az")})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	a, err := az.New(l, c)
	require.NoError(t, err)

	op, err := onepassword.New(l, c)
	require.NoError(t, err)

	cmd, err := pulumi.NewCommand(l, a, op, c)
	require.NoError(t, err)

	testutils.AssertDescribed(t, "pulumi", cmd)
	testutils.AssertSkilled(t, "pulumi", cmd)
}
