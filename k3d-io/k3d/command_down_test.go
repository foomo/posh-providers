package k3d_test

import (
	"os"
	"path"
	"strings"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/k3d-io/k3d"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDownSharedRegistry covers the teardown of the registry every cluster
// shares.
//
// `down` used to delete the registry whenever it existed - the author's own
// "TODO check if empty" sat above the line. Config holds one Registry against a
// Clusters map, so with two clusters up, `k3d down local` left `test` running
// against a registry that no longer existed.
//
// A stub `k3d` earlier on PATH answers `cluster list` / `registry list` with
// canned JSON and records every invocation, so this asserts on what the provider
// actually ran.
func TestDownSharedRegistry(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	tests := []struct {
		name          string
		running       []string // clusters `k3d cluster list` reports
		deleteCluster string
		wantDeleted   bool
	}{
		{
			name:          "another cluster still up keeps the registry",
			running:       []string{"local", "test"},
			deleteCluster: "local",
			wantDeleted:   false,
		},
		{
			name:          "last cluster removes the registry",
			running:       []string{"local"},
			deleteCluster: "local",
			wantDeleted:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := stubK3d(t, tt.running, tt.deleteCluster)

			viper.Set("kubectl", map[string]any{"configPath": t.TempDir()})
			viper.Set("k3d", map[string]any{
				"registry": map[string]any{"name": "shared-registry", "port": "12345"},
				"clusters": map[string]any{
					"local": map[string]any{},
					"test":  map[string]any{},
				},
			})
			t.Cleanup(viper.Reset)

			l, c := log.NewFmt(), &cache.MemoryCache{}

			kc, err := kubectl.New(l, c)
			require.NoError(t, err)

			kd, err := k3d.New(l)
			require.NoError(t, err)

			cmd, err := k3d.NewCommand(l, kd, c, kc)
			require.NoError(t, err)

			r, err := readline.New(l)
			require.NoError(t, err)
			require.NoError(t, r.Parse("k3d down "+tt.deleteCluster))
			require.NoError(t, cmd.Execute(t.Context(), r))

			got, err := os.ReadFile(calls)
			require.NoError(t, err)

			assert.Contains(t, string(got), "cluster delete "+tt.deleteCluster,
				"the named cluster must always be deleted")

			if tt.wantDeleted {
				assert.Contains(t, string(got), "registry delete shared-registry",
					"the last cluster going down must remove the shared registry")
			} else {
				assert.NotContains(t, string(got), "registry delete",
					"the shared registry must survive while another cluster uses it")
			}
		})
	}
}

// stubK3d installs a fake `k3d` on PATH that reports the given clusters and an
// existing registry, and records each invocation. It returns the record path.
//
// `cluster list` is asked twice with different intent: once for the cluster
// being deleted (before the delete) and once for the survivors (after), so the
// stub drops the deleted cluster from its answer as soon as the delete has run.
func stubK3d(t *testing.T, running []string, deleted string) string {
	t.Helper()

	bin := t.TempDir()
	calls := path.Join(bin, "calls.txt")
	deletedMarker := path.Join(bin, "deleted")

	var before, after []string

	for _, name := range running {
		before = append(before, `{"name":"`+name+`"}`)

		if name != deleted {
			after = append(after, `{"name":"`+name+`"}`)
		}
	}

	stub := `#!/bin/sh
echo "$*" >> ` + calls + `
case "$1 $2" in
"cluster list")
  if [ -f ` + deletedMarker + ` ]; then
    echo '[` + strings.Join(after, ",") + `]'
  else
    echo '[` + strings.Join(before, ",") + `]'
  fi
  ;;
"registry list")
  echo '[{"name":"k3d-shared-registry"}]'
  ;;
"cluster delete")
  touch ` + deletedMarker + `
  ;;
esac
exit 0
`
	require.NoError(t, os.WriteFile(path.Join(bin, "k3d"), []byte(stub), 0o700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	return calls
}
