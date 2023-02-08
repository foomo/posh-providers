package gcloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestGCloud_ParseAccounts(t *testing.T) {
	gcloud := GCloud{
		cfg: Config{
			ConfigDir: "testdata/accounts",
		},
	}
	accounts, err := gcloud.ParseAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, accounts, 1)

	assert.Equal(t, "testdata/accounts/admin@prod-default.json", accounts[0].Path)
	assert.Equal(t, "admin", accounts[0].Role)
	assert.Equal(t, "prod", accounts[0].Environment)
	assert.Equal(t, "default", accounts[0].Cluster)
}
