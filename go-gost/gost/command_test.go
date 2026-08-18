package gost_test

import (
	"log/slog"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh-providers/go-gost/gost"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Not parallel: NewCommand reads its config through viper, which is a singleton.
	l := log.NewFmt()

	cmd, err := gost.NewCommand(l, gokazi.New(slog.New(l.SlogHandler())))
	require.NoError(t, err)

	testutils.AssertDescribed(t, "gost", cmd)
	testutils.AssertSkilled(t, "gost", cmd)
}
