package rclone_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh-providers/rclone/rclone"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// No t.Parallel: viper is a singleton and NewCommand exports RCLONE_CONFIG.
func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	viper.Set("rclone", map[string]any{"path": path.Join(dir, "rclone.conf")})
	t.Cleanup(viper.Reset)
	t.Setenv("RCLONE_CONFIG", os.Getenv("RCLONE_CONFIG"))

	cmd, err := rclone.NewCommand(log.NewFmt(), &cache.MemoryCache{})
	require.NoError(t, err)

	testutils.AssertDescribed(t, "rclone", cmd)
	testutils.AssertSkilled(t, "rclone", cmd)
}
