package terraform_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/hashicorp/terraform"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// getWorkspaces walks the configured path, so give it a real workspace.
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(path.Join(dir, "staging"), 0700))
	require.NoError(t, os.WriteFile(path.Join(dir, "staging", "main.tf"), []byte("resource \"null_resource\" \"a\" {}\n"), 0600))

	viper.Set("terraform", map[string]any{
		"path":     dir,
		"tenantId": "00000000-0000-0000-0000-000000000000",
		"subscriptions": map[string]any{
			"staging": map[string]any{
				"id": "00000000-0000-0000-0000-000000000000",
				"backend": map[string]any{
					"resourceGroupName":  "rg",
					"storageAccountName": "sa",
					"containerName":      "tfstate",
				},
			},
		},
		"servicePrincipals": map[string]any{
			"ci": map[string]any{
				"tenantId":       "00000000-0000-0000-0000-000000000000",
				"clientId":       "00000000-0000-0000-0000-000000000000",
				"clientSecret":   "secret",
				"subscriptionId": "00000000-0000-0000-0000-000000000000",
			},
		},
	})
	t.Cleanup(viper.Reset)

	cmd, err := terraform.NewCommand(log.NewFmt(), &cache.MemoryCache{})
	require.NoError(t, err)

	testutils.AssertDescribed(t, "terraform", cmd)
	testutils.AssertSkilled(t, "terraform", cmd)
}
