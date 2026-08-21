package teleport_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/gravitational/teleport"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKubeconfigPreservesConfigOnFailure covers what happens to an existing
// kubeconfig when the login that should replace it fails.
//
// `kubeconfig` used to call cluster.DeleteConfig (a real os.Remove) and only
// then run `tsh kube login`, so an expired session or a network failure left the
// profile with no kubeconfig at all - having destroyed a working one. The delete
// itself is load-bearing: upstream's kubeconfig.Update calls LoadConfig on the
// path and *merges*, so leaving the old file in place would accumulate stale
// contexts. Hence move-aside-and-restore rather than simply reordering.
//
// A stub `tsh` earlier on PATH decides whether the login succeeds.
func TestKubeconfigPreservesConfigOnFailure(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	const existing = "apiVersion: v1\n# the working config\n"

	tests := []struct {
		name      string
		exitCode  string
		wantErr   bool
		wantAfter string
	}{
		{
			name:      "a failed login leaves the old config in place",
			exitCode:  "1",
			wantErr:   true,
			wantAfter: existing,
		},
		{
			name:      "a successful login replaces it",
			exitCode:  "0",
			wantErr:   false,
			wantAfter: "written by tsh\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := t.TempDir()
			viper.Set("kubectl", map[string]any{"configPath": configPath})
			viper.Set("teleport", map[string]any{"path": t.TempDir()})
			t.Cleanup(viper.Reset)

			// The kubeconfig the profile already has.
			profileDir := path.Join(configPath, "teleport")
			require.NoError(t, os.MkdirAll(profileDir, 0o700))
			target := path.Join(profileDir, "mycluster.yaml")
			require.NoError(t, os.WriteFile(target, []byte(existing), 0o600))

			bin := t.TempDir()
			stub := "#!/bin/sh\necho 'written by tsh' > \"$KUBECONFIG\"\nexit " + tt.exitCode + "\n"
			require.NoError(t, os.WriteFile(path.Join(bin, "tsh"), []byte(stub), 0o700))
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

			l, c := log.NewFmt(), &cache.MemoryCache{}

			kc, err := kubectl.New(l, c)
			require.NoError(t, err)

			tp, err := teleport.NewTeleport(l, c)
			require.NoError(t, err)

			cmd := teleport.NewCommand(l, c, tp, kc)

			r, err := readline.New(l)
			require.NoError(t, err)
			require.NoError(t, r.Parse("teleport kubeconfig mycluster --profile teleport"))

			err = cmd.Execute(t.Context(), r)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			got, err := os.ReadFile(target)
			require.NoError(t, err, "the kubeconfig must still exist either way")
			assert.Equal(t, tt.wantAfter, string(got))

			// No stash left behind in either case.
			entries, err := os.ReadDir(profileDir)
			require.NoError(t, err)

			for _, entry := range entries {
				assert.NotContains(t, entry.Name(), "posh-stash",
					"the stashed copy must be cleaned up")
			}
		})
	}
}
