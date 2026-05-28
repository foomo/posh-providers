package golang

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestPathsIgnoresConfiguredDirectories(t *testing.T) {
	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "packages", "go", "go.mod"))
	mustWrite(t, filepath.Join(tmp, ".worktrees", "feature", "packages", "go", "go.mod"))

	paths, err := findPaths(context.Background(), tmp, "go.mod", true, []string{`^\.worktrees$`})
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(tmp, "packages", "go")
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
