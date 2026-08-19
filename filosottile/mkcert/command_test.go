package mkcert_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/filosottile/mkcert"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	viper.Set("mkcert", map[string]any{
		"certificatePath": path.Join(t.TempDir(), "certs"),
		"certificates": []map[string]any{
			{"name": "foomo.org", "names": []string{"foomo.org", "*.foomo.org", "localhost"}},
		},
	})
	t.Cleanup(viper.Reset)

	cmd, err := mkcert.NewCommand(log.NewFmt())
	require.NoError(t, err)

	testutils.AssertDescribed(t, "mkcert", cmd)
	testutils.AssertSkilled(t, "mkcert", cmd)
}
