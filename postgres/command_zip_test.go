package postgres_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/postgres"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithoutZipProvider covers constructing the command without
// CommandWithZip.
//
// The option is presented as optional, but `dump` and `restore` called
// inst.zip.Config() while *building their flag sets* - so completing or
// executing either verb dereferenced a nil *zip.Zip and panicked. The README's
// snippet is postgres.NewCommand(l), so a project wired from it crashed on first
// use.
//
// Not parallel: NewCommand reads config through viper, which is a singleton.
func TestWithoutZipProvider(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	viper.Set("postgres", map[string]any{
		"dumpPath": path.Join(t.TempDir(), "dumps"),
		"databases": map[string]any{
			"local": map[string]any{"host": "localhost", "port": 5432, "user": "postgres"},
		},
	})
	t.Cleanup(viper.Reset)

	cmd := postgres.NewCommand(log.NewFmt())

	// Completion builds the flag sets, which is where the nil deref used to fire.
	for _, line := range []string{
		"postgres dump ",
		"postgres restore ",
		"postgres dump local --",
	} {
		t.Run(line, func(t *testing.T) {
			r, err := readline.New(log.NewFmt())
			require.NoError(t, err)
			require.NoError(t, r.Parse(line))

			assert.NotPanics(t, func() {
				_ = cmd.Complete(t.Context(), r)
			}, "completion must not require the zip provider")
		})
	}
}
