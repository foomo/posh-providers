package gcloud_test

import (
	"context"
	"testing"

	"github.com/foomo/posh-providers/google/gcloud"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGCloud_ParseAccounts(t *testing.T) {
	l := log.NewTest(t)
	c := &cache.MemoryCache{}
	inst, err := gcloud.New(l, c, gcloud.CommandWithConfig(
		&gcloud.Config{
			ConfigDir:    "testdata/accounts",
			Environments: nil,
		},
	))
	require.NoError(t, err)

	accounts, err := inst.ParseAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, accounts, 1)

	assert.Equal(t, "testdata/accounts/admin@prod-default.json", accounts[0].Path)
	assert.Equal(t, "admin", accounts[0].Role)
	assert.Equal(t, "prod", accounts[0].Environment)
	assert.Equal(t, "default", accounts[0].Cluster)
}
