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

// TestStopFindsEitherProcess covers stopping a dockprox started by either verb.
//
// The provider registered a single gokazi task carrying the config path as an
// arg, and gokazi matches a running process by requiring *every* registered arg
// to appear in its command line. `menubar` runs a bare `dockprox menubar` with no
// config path, so it never matched: `stop` reported nothing running, and because
// the ErrAlreadyRunning guard uses the same match, `start` afterwards launched a
// second dockprox on the same ports.
//
// The two verbs now register separate task ids and `stop` tries both. With
// neither running, both report ErrNotRunning - which is expected, not a failure,
// so stop must succeed quietly rather than erroring.
//
// Not parallel: NewCommand reads config through viper, which is a singleton.
func TestStopFindsEitherProcess(t *testing.T) {
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

	// Nothing is running, so both registered tasks report ErrNotRunning. That is
	// the normal case and must not surface as an error.
	assert.NoError(t, cmd.Execute(t.Context(), r),
		"stopping with nothing running must not be an error")
}

// TestBothTasksRegistered pins that each verb has its own gokazi task, which is
// what lets stop find a menubar process.
func TestBothTasksRegistered(t *testing.T) {
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
	assert.Contains(t, tasks, "dockprox.menubar",
		"menubar needs its own task or stop cannot find the process it starts")
}
