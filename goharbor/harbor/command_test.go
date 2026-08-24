package harbor_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/goharbor/harbor"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	viper.Set("harbor", map[string]any{
		"url":     "https://harbor.example.org",
		"authUrl": "https://harbor.example.org/c/oidc/login",
		"project": "example",
	})
	t.Cleanup(viper.Reset)

	h, err := harbor.New(log.NewFmt())
	require.NoError(t, err)

	cmd := harbor.NewCommand(log.NewFmt(), h)

	testutils.AssertDescribed(t, "harbor", cmd)
	testutils.AssertSkilled(t, "harbor", cmd)
}
