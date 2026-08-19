package cdktf_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/hashicorp/cdktf"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// stacks() walks the configured path for *.stack.yaml files.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(path.Join(dir, "my.stack.yaml"), []byte("{}\n"), 0600))

	viper.Set("cdktf", map[string]any{
		"path":    dir,
		"outPath": path.Join(dir, "cdktf.out"),
	})
	t.Cleanup(viper.Reset)

	cmd, err := cdktf.NewCommand(log.NewFmt(), &cache.MemoryCache{})
	require.NoError(t, err)

	testutils.AssertDescribed(t, "cdktf", cmd)
	testutils.AssertSkilled(t, "cdktf", cmd)
}
