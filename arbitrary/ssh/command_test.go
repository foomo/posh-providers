package ssh_test

import (
	"log/slog"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh-providers/arbitrary/ssh"
	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/foomo/posh/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestCommandDocs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	// Not parallel: ssh.New reads its config through viper, which is a singleton.
	l := log.NewFmt()

	s, err := ssh.New(l, gokazi.New(slog.New(l.SlogHandler())))
	require.NoError(t, err)

	cmd := ssh.NewCommand(l, s)

	testutils.AssertDescribed(t, "ssh", cmd)
	testutils.AssertSkilled(t, "ssh", cmd)
}
