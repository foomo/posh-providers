package cloudflared_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/cloudflare/cloudflared"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDescribe(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// NewCommand does os.MkdirAll on Config().Path, so the config key must be
	// set before constructing - an unset path fails with "mkdir : no such file
	// or directory". Verified working.
	viper.Set("cloudflared", map[string]any{"path": t.TempDir()})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()

	cf, err := cloudflared.New(l)
	require.NoError(t, err)

	cmd, err := cloudflared.NewCommand(l, cf)
	require.NoError(t, err)

	testutils.AssertDescribed(t, "cloudflared", cmd)
	testutils.AssertSkilled(t, "cloudflared", cmd)
}
