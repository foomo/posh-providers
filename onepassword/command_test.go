package onepassword_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	viper.Set("onePassword", map[string]any{
		"account":       "example",
		"tokenFilename": path.Join(t.TempDir(), ".op"),
	})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()

	op, err := onepassword.New(l, &cache.MemoryCache{})
	require.NoError(t, err)

	cmd, err := onepassword.NewCommand(l, op)
	require.NoError(t, err)

	testutils.AssertDescribed(t, "op", cmd)
	testutils.AssertSkilled(t, "op", cmd)
}
