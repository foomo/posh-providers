package gokazi_test

import (
	"log/slog"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	gk "github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh-providers/foomo/gokazi"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	l := log.NewFmt()

	cmd, err := gokazi.NewCommand(l, gk.New(slog.New(l.SlogHandler())))
	require.NoError(t, err)

	testutils.AssertDescribed(t, "gokazi", cmd)
	testutils.AssertSkilled(t, "gokazi", cmd)
}
