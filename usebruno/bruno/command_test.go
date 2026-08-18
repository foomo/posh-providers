package bruno_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/pkg/agentdoc"
	"github.com/foomo/posh-providers/usebruno/bruno"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	cmd, err := bruno.NewCommand(log.NewFmt())
	require.NoError(t, err)

	agentdoc.AssertDescribed(t, "bruno", cmd)
}
