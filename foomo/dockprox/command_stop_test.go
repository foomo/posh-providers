package dockprox_test

import (
	"log/slog"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh-providers/foomo/dockprox"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStopWithNothingRunning covers stopping when no proxy is up.
//
// The task is always registered, so gokazi reports ErrNotRunning - which is
// expected, not a failure, so stop must succeed quietly rather than erroring.
//
// Not parallel: NewCommand reads config through viper, which is a singleton.
func TestStopWithNothingRunning(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	viper.Set("dockprox", map[string]any{"config": "devops/dockprox.yaml"})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()
	gk := gokazi.New(slog.New(l.SlogHandler()))

	cmd, err := dockprox.NewCommand(l, gk)
	require.NoError(t, err)

	r, err := readline.New(l)
	require.NoError(t, err)
	require.NoError(t, r.Parse("dockprox stop"))

	// Nothing is running, so the registered task reports ErrNotRunning. That is
	// the normal case and must not surface as an error.
	assert.NoError(t, cmd.Execute(t.Context(), r),
		"stopping with nothing running must not be an error")
}

// TestServeTaskRegistered pins the gokazi task id the proxy is registered under.
func TestServeTaskRegistered(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	viper.Set("dockprox", map[string]any{"config": "devops/dockprox.yaml"})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()
	gk := gokazi.New(slog.New(l.SlogHandler()))

	_, err := dockprox.NewCommand(l, gk)
	require.NoError(t, err)

	tasks, err := gk.List(t.Context())
	require.NoError(t, err)

	assert.Contains(t, tasks, "dockprox.serve")
}
