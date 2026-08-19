package webdriverio_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh-providers/webdriverio/webdriverio"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Not parallel: NewCommand reads its config through viper, which is a singleton.
	l := log.NewFmt()

	op, err := onepassword.New(l, &cache.MemoryCache{})
	require.NoError(t, err)

	cmd, err := webdriverio.NewCommand(l, &cache.MemoryCache{}, op)
	require.NoError(t, err)

	testutils.AssertDescribed(t, "wdio", cmd)
	testutils.AssertSkilled(t, "wdio", cmd)
}
