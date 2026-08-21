package licensefinder_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/pivotal/licensefinder"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAggregatePaths covers how multiple source directories reach license_finder.
//
// The provider used to pass one joined argument, `--aggregate_paths=a b c`. posh
// runs everything through `sh -c`, so the shell split that on the spaces, and
// Thor - which declares the option `type: :array` - stops collecting at the value
// attached with `=`. Only the first path was aggregated; the rest arrived as
// positionals the verbs do not accept. Upstream's own help spells the form as
// `--aggregate_paths='a' 'b'`.
//
// A stub `license_finder` earlier on PATH records the argv it was called with.
func TestAggregatePaths(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	for _, project := range []string{"one", "two"} {
		require.NoError(t, os.MkdirAll(path.Join(dir, project), 0o700))
		require.NoError(t, os.WriteFile(path.Join(dir, project, "go.mod"),
			[]byte("module "+project+"\n"), 0o600))
	}

	t.Chdir(dir)
	t.Setenv("PROJECT_ROOT", dir)

	bin := t.TempDir()
	record := path.Join(bin, "argv.txt")
	stub := "#!/bin/sh\necho \"$*\" > " + record + "\nexit 0\n"
	require.NoError(t, os.WriteFile(path.Join(bin, "license_finder"), []byte(stub), 0o700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	viper.Set("licensefinder", map[string]any{
		"sources":       []string{"go.mod"},
		"logPath":       path.Join(dir, "log"),
		"decisionsPath": path.Join(dir, "decisions.yml"),
	})
	t.Cleanup(viper.Reset)

	cmd, err := licensefinder.NewCommand(log.NewFmt(), &cache.MemoryCache{})
	require.NoError(t, err)

	r, err := readline.New(log.NewFmt())
	require.NoError(t, err)
	require.NoError(t, r.Parse("licensefinder report"))
	require.NoError(t, cmd.Execute(t.Context(), r))

	out, err := os.ReadFile(record)
	require.NoError(t, err)

	got := string(out)

	// The shell strips the quotes, so both directories arrive as separate tokens
	// after the flag name rather than glued to it with '='.
	assert.Contains(t, got, "--aggregate_paths one two",
		"every discovered path must follow the flag as its own argument")
	assert.NotContains(t, got, "--aggregate_paths=",
		"the = form makes Thor stop collecting after the first path")

	// Sanity: both projects really were discovered.
	assert.Contains(t, got, "one")
	assert.Contains(t, got, "two")
}
