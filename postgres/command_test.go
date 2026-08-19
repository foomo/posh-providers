package postgres_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/arbitrary/zip"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh-providers/postgres"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// CommandWithZip is mandatory in practice: dump/restore dereference the zip
	// provider while building their flags. Not parallel: zip.New reads viper.
	l := log.NewFmt()

	op, err := onepassword.New(l, &cache.MemoryCache{})
	require.NoError(t, err)

	z, err := zip.New(l, op)
	require.NoError(t, err)

	cmd := postgres.NewCommand(l, postgres.CommandWithZip(z))

	testutils.AssertDescribed(t, "postgres", cmd)
	testutils.AssertSkilled(t, "postgres", cmd)
}
