package cli_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/cli/cli"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	l := log.NewFmt()

	c, err := cli.New(l)
	require.NoError(t, err)

	cmd := cli.NewCommand(l, c)

	testutils.AssertDescribed(t, "gh", cmd)
	testutils.AssertSkilled(t, "gh", cmd)
}
