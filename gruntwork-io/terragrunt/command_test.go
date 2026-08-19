package terragrunt_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/gruntwork-io/terragrunt"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Envs, sites and stacks are all discovered from the filesystem.
	dir := t.TempDir()
	stack := path.Join(dir, "envs", "prod", "eu", "net", "vpc")
	require.NoError(t, os.MkdirAll(stack, 0700))
	require.NoError(t, os.WriteFile(path.Join(stack, "terragrunt.hcl"), []byte("{}\n"), 0600))

	viper.Set("terragrunt", map[string]any{
		"path":      dir,
		"cachePath": path.Join(dir, "cache"),
	})
	viper.Set("onePassword", map[string]any{"account": "example"})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	op, err := onepassword.New(l, c)
	require.NoError(t, err)

	cmd, err := terragrunt.NewCommand(l, op, c)
	require.NoError(t, err)

	testutils.AssertDescribed(t, "terragrunt", cmd)
	testutils.AssertSkilled(t, "terragrunt", cmd)
}
