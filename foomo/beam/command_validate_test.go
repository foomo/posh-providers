package beam_test

import (
	"log/slog"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh-providers/cloudflare/cloudflared"
	"github.com/foomo/posh-providers/foomo/beam"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidate covers rejecting names that are not in the config.
//
// GetCluster/GetDatabase index the config map and return a zero value rather
// than an error, so an unknown name used to reach `connect` as hostname "" and
// port 0, and `disconnect` as a `--hostname ` substring matching every running
// cloudflared process. The ClusterExists/DatabaseExists helpers existed for this
// and had no callers.
//
// Not parallel: viper is a singleton and the constructors mutate the process
// environment.
func TestValidate(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	viper.Set("kubectl", map[string]any{"configPath": path.Join(dir, "kubectl")})
	viper.Set("beam", map[string]any{
		"clusters": map[string]any{
			"prod": map[string]any{"port": 12200, "hostname": "prod.example.com"},
		},
		"databases": map[string]any{
			"main": map[string]any{"port": 12202, "hostname": "main.example.com"},
		},
	})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	op, err := onepassword.New(l, c)
	require.NoError(t, err)

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	cf, err := cloudflared.New(l)
	require.NoError(t, err)

	b, err := beam.New(l, op, gokazi.New(slog.New(l.SlogHandler())))
	require.NoError(t, err)

	cmd, err := beam.NewCommand(l, b, kc, cf)
	require.NoError(t, err)

	tests := []struct {
		line    string
		wantErr string
	}{
		{"beam cluster connect prod", ""},
		{"beam cluster connect prodd", "invalid [cluster] argument: prodd"},
		{"beam cluster kubeconfig prod", ""},
		{"beam cluster kubeconfig nope", "invalid [cluster] argument: nope"},
		{"beam database connect main", ""},
		{"beam database connect mainn", "invalid [database] argument: mainn"},
		// A database name is not a cluster name and vice versa.
		{"beam cluster connect main", "invalid [cluster] argument: main"},
		{"beam database connect prod", "invalid [database] argument: prod"},
		// Omitting the name on disconnect deliberately means "all of them".
		{"beam cluster disconnect", ""},
		{"beam database disconnect", ""},
		// A verb taking no name at all.
		{"beam status", ""},
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
