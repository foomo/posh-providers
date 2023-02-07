package gcloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAccounts(t *testing.T) {
	accounts, err := ParseAccounts(context.Background(), "testdata/accounts")
	require.NoError(t, err)
	require.Len(t, accounts, 2)

	assert.Equal(t, "testdata/accounts/admin@prod-default.json", accounts[0].Path)
	assert.Equal(t, "admin", accounts[0].Role)
	assert.Equal(t, "prod", accounts[0].Environment)
	assert.Equal(t, "default", accounts[0].Cluster)
}
