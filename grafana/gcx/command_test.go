package gcx_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/grafana/gcx"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Not parallel: NewCommand reads its config through viper, which is a singleton.
	cmd, err := gcx.NewCommand(log.NewFmt(), &cache.MemoryCache{})
	require.NoError(t, err)

	testutils.AssertDescribed(t, "gcx", cmd)
	testutils.AssertSkilled(t, "gcx", cmd)
}
