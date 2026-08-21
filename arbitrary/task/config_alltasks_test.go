package task_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/arbitrary/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAllTasks covers merging file tasks with configured ones.
//
// AllTasks used to do `ret := c.Tasks`, which aliases the map rather than
// copying it, then wrote the file tasks into it. Two consequences: a file task
// silently overwrote the caller's *configured* task of the same name, and a nil
// Tasks map with a Path set panicked with "assignment to entry in nil map".
func TestAllTasks(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	t.Run("does not mutate the configured tasks", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.WriteFile(path.Join(dir, "build.yaml"),
			[]byte("description: from file\ncmds:\n  - echo file\n"), 0o600))

		cfg := task.Config{
			Path: dir,
			Tasks: map[string]task.Task{
				"build": {Description: "from config"},
			},
		}

		all, err := cfg.AllTasks()
		require.NoError(t, err)

		assert.Equal(t, "from file", all["build"].Description,
			"a file task still wins on a name collision")
		assert.Equal(t, "from config", cfg.Tasks["build"].Description,
			"the caller's config must be left alone")
	})

	t.Run("nil task map with a path does not panic", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.WriteFile(path.Join(dir, "x.yaml"),
			[]byte("cmds:\n  - echo x\n"), 0o600))

		cfg := task.Config{Path: dir} // Tasks is nil

		all, err := cfg.AllTasks()
		require.NoError(t, err)
		assert.Contains(t, all, "x")
	})

	t.Run("file tasks are added alongside configured ones", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.WriteFile(path.Join(dir, "fromfile.yaml"),
			[]byte("cmds:\n  - echo x\n"), 0o600))

		cfg := task.Config{
			Path:  dir,
			Tasks: map[string]task.Task{"fromconfig": {Description: "c"}},
		}

		all, err := cfg.AllTasks()
		require.NoError(t, err)
		assert.Contains(t, all, "fromfile")
		assert.Contains(t, all, "fromconfig")
	})
}
