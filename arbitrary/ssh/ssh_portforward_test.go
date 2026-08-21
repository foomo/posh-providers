package ssh_test

import (
	"log/slog"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh-providers/arbitrary/ssh"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPortForwardAutoAssignsPort covers `port: 0` meaning "pick a free port".
//
// StartPortForward computed a free port when the configured one was <= 0, but
// then built the -L flag from the raw config value - so a 0 reached ssh verbatim
// and the computed value was only used in the "ready at" log line, naming a port
// nothing was listening on. StartSocks5Tunnel used the computed value and was
// correct.
//
// A stub `ssh` earlier on PATH records the argv it was called with.
func TestPortForwardAutoAssignsPort(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	bin := t.TempDir()
	record := path.Join(bin, "argv.txt")
	// Record argv, then exit: a lingering process would make gokazi report
	// ErrAlreadyRunning on a later run and skip the launch entirely, which made
	// this test order-dependent.
	stub := "#!/bin/sh\necho \"$*\" > " + record + "\n"
	require.NoError(t, os.WriteFile(path.Join(bin, "ssh"), []byte(stub), 0o700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	viper.Set("ssh", map[string]any{
		"portForwards": map[string]any{
			"db": map[string]any{
				"port":     0, // auto-assign
				"host":     "db.internal",
				"hostPort": 5432,
				"username": "user",
			},
		},
	})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()

	inst, err := ssh.New(l, gokazi.New(slog.New(l.SlogHandler())))
	require.NoError(t, err)

	// The forward is started in the background, so poll for the stub's record
	// rather than assuming it has been written by the time Start returns.
	_ = inst.StartPortForward(t.Context(), "db")

	var out []byte

	require.Eventually(t, func() bool {
		var err error

		out, err = os.ReadFile(record)

		return err == nil && len(out) > 0
	}, 5*time.Second, 20*time.Millisecond, "the stub should have been invoked")

	got := string(out)
	assert.NotContains(t, got, "-L 0:", "a configured port of 0 must not reach ssh verbatim")
	assert.Contains(t, got, ":db.internal:5432", "the forward target must be preserved")

	// The bound port must be a real, non-zero one.
	idx := strings.Index(got, "-L ")
	require.GreaterOrEqual(t, idx, 0, "the -L flag must be present")

	spec := strings.Fields(got[idx+len("-L "):])[0]
	local := strings.SplitN(spec, ":", 2)[0]
	assert.NotEqual(t, "0", local, "the computed free port must be used, not the configured 0")
}
