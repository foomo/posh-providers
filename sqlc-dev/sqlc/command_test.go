package sqlc_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/foomo/posh-providers/sqlc-dev/sqlc"
)

func TestPathsIgnoresConfiguredDirectories(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "packages", "go", "domain", "loan", "sqlc.yaml"))
	mustWrite(t, filepath.Join(tmp, ".worktrees", "feature", "packages", "go", "domain", "loan", "sqlc.yaml"))

	paths, err := sqlc.FindPaths(context.Background(), tmp, []string{`^\.worktrees$`})
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(tmp, "packages", "go", "domain", "loan", "sqlc.yaml")
	if !slices.Equal(paths, []string{want}) {
		t.Fatalf("paths = %v, want %v", paths, []string{want})
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
}
