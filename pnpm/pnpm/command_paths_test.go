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

// TestWorkspacePathCompletion covers path discovery when there is no
// pnpm-workspace.yaml.
//
// The missing-file branch was guarded with errors.Is(err, os.ErrExist), which
// os.Stat never returns - so an absent workspace file fell into the next branch
// (err != nil) and returned an empty list instead of reaching the roots = ["."]
// fallback below it. In a single-package project the whole `workspace` subtree
// therefore offered no paths at all.
//
// Not parallel: this changes the working directory and PROJECT_ROOT.
func TestWorkspacePathCompletion(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	t.Run("single package project falls back to the root", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(path.Join(dir, "package.json"),
			[]byte(`{"name":"x","scripts":{"build":"tsc"}}`), 0o600))
		t.Chdir(dir)
		t.Setenv("PROJECT_ROOT", dir)

		got := complete(t, "pnpm workspace ")
		assert.Contains(t, got, ".", "with no pnpm-workspace.yaml the project root must still be offered")
	})

	t.Run("workspace file lists its packages", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(path.Join(dir, "pnpm-workspace.yaml"),
			[]byte("packages:\n  - 'pkg/*'\n"), 0o600))
		require.NoError(t, os.MkdirAll(path.Join(dir, "pkg", "one"), 0o700))
		require.NoError(t, os.WriteFile(path.Join(dir, "pkg", "one", "package.json"),
			[]byte(`{"name":"one"}`), 0o600))
		t.Chdir(dir)
		t.Setenv("PROJECT_ROOT", dir)

		got := complete(t, "pnpm workspace ")
		assert.Contains(t, got, "pkg/one")
	})
}

func complete(t *testing.T, line string) []string {
	t.Helper()

	cmd := pnpm.NewCommand(log.NewFmt(), &cache.MemoryCache{})

	r, err := readline.New(log.NewFmt())
	require.NoError(t, err)
	require.NoError(t, r.Parse(line))

	var ret []string
	for _, s := range cmd.Complete(t.Context(), r) {
		ret = append(ret, s.Text)
	}

	return ret
}
