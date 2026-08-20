package k9s_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/derailed/k9s"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidate covers rejecting a cluster with no kubeconfig behind it.
//
// The cluster is a tree.Arg with a Suggest, and args are not matched against
// their suggestions - so a typo reaches Execute, where it became a KUBECONFIG
// pointing at a missing file and a confusing downstream error. It must be
// rejected up front, and the --profile flag has to be honoured while doing it:
// a cluster can exist only under a profile.
//
// Not parallel: NewCommand reads config through viper, which is a singleton.
func TestValidate(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(path.Join(dir, "staging"), 0o700))
	require.NoError(t, os.WriteFile(path.Join(dir, "prod.yaml"), []byte("apiVersion: v1\n"), 0o600))
	require.NoError(t, os.WriteFile(path.Join(dir, "staging", "eu.yaml"), []byte("apiVersion: v1\n"), 0o600))

	viper.Set("kubectl", map[string]any{"configPath": dir})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	cmd := k9s.NewCommand(l, kc)

	tests := []struct {
		line    string
		wantErr string
	}{
		{"k9s prod", ""},
		{"k9s prodd", "invalid [cluster] argument: prodd"},
		{"k9s", "missing [cluster] argument"},
		// eu exists only under the staging profile
		{"k9s eu --profile staging", ""},
		{"k9s eu", "invalid [cluster] argument: eu"},
		{"k9s euu --profile staging", "invalid [cluster] argument: euu"},
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
