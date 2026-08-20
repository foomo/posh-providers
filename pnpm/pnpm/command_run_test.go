package pnpm_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/pnpm/pnpm"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunForwardsFlags covers flags reaching pnpm from the `run` leaves.
//
// The run helper took no readline, so it could not forward anything: every
// other leaf in this tree forwards r.Flags(), but `pnpm run build --silent`
// dropped --silent without a word.
//
// A stub `pnpm` earlier on PATH records the argv it was called with, so this
// asserts the provider's own call site rather than shell's semantics.
//
// Not parallel: this changes the working directory, PATH and PROJECT_ROOT.
func TestRunForwardsFlags(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(path.Join(dir, "package.json"),
		[]byte(`{"name":"x","scripts":{"build":"tsc"}}`), 0o600))
	t.Chdir(dir)
	t.Setenv("PROJECT_ROOT", dir)

	bin := t.TempDir()
	record := path.Join(bin, "argv.txt")
	stub := "#!/bin/sh\necho \"$*\" > " + record + "\n"
	require.NoError(t, os.WriteFile(path.Join(bin, "pnpm"), []byte(stub), 0o700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	cmd := pnpm.NewCommand(log.NewFmt(), &cache.MemoryCache{})

	r, err := readline.New(log.NewFmt())
	require.NoError(t, err)
	require.NoError(t, r.Parse("pnpm run build --recursive"))
	require.NoError(t, cmd.Execute(t.Context(), r))

	out, err := os.ReadFile(record)
	require.NoError(t, err)

	assert.Equal(t, "run build --recursive\n", string(out),
		"the script name and the typed flag must both reach pnpm")
}
