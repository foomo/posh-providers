package docusaurus_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/facebook/docusaurus"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	cmd, err := docusaurus.NewCommand(log.NewFmt())
	require.NoError(t, err)

	testutils.AssertDescribed(t, "docusaurus", cmd)
	testutils.AssertSkilled(t, "docusaurus", cmd)
}
