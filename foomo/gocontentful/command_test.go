package gocontentful_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/foomo/gocontentful"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Not parallel: onepassword.New reads its config through viper, a singleton.
	l, c := log.NewFmt(), &cache.MemoryCache{}

	op, err := onepassword.New(l, c)
	require.NoError(t, err)

	cmd := gocontentful.NewCommand(l, c, op)

	testutils.AssertDescribed(t, "gocontentful", cmd)
	testutils.AssertSkilled(t, "gocontentful", cmd)
}
