package helm_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/helm/helm"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateProfile covers validating the cluster against the kubeconfig that
// execution will actually read.
//
// Validate used to call ConfigExists(""), always checking the top-level
// <configPath>/<cluster>.yaml, while execute reads
// <configPath>/<profile>/<cluster>.yaml. So a cluster that existed only under a
// profile was rejected outright, and one that existed only at top level passed
// validation and then failed inside helm.
//
// Not parallel: kubectl.New mutates the process environment and viper is a
// singleton.
func TestValidateProfile(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(path.Join(dir, "azure"), 0o700))
	require.NoError(t, os.WriteFile(path.Join(dir, "top.yaml"), []byte("apiVersion: v1\n"), 0o600))
	require.NoError(t, os.WriteFile(path.Join(dir, "azure", "scoped.yaml"), []byte("apiVersion: v1\n"), 0o600))

	viper.Set("kubectl", map[string]any{"configPath": dir})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	cmd := helm.NewCommand(l, kc)

	tests := []struct {
		line    string
		wantErr string
	}{
		// top.yaml exists only at the top level
		{"helm top list", ""},
		{"helm top list --profile azure", "invalid [CLUSTER] argument"},
		// scoped.yaml exists only under the azure profile
		{"helm scoped list --profile azure", ""},
		{"helm scoped list", "invalid [CLUSTER] argument"},
		// neither
		{"helm nope list", "invalid [CLUSTER] argument"},
		{"helm", "missing [CLUSTER] argument"},
		{"helm top", "missing [CMD] argument"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			r, err := readline.New(l)
			require.NoError(t, err)
			require.NoError(t, r.Parse(tt.line))

			err = cmd.Validate(t.Context(), r)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
