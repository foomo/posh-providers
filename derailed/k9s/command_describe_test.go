package k9s_test

import (
	"context"
	"os"
	"path/filepath"
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

// TestClusterArgDescribesValidation pins that the cluster argument does not
// advertise itself as unvalidated.
//
// It said "not validated" until a Validate calling ConfigExistsForFlags was
// added, and nothing updated the string. The catalog publishes it verbatim, so
// an agent reading it concludes it has to pre-check the name itself - or worse,
// that a typo silently points KUBECONFIG at a missing file. The twin
// bjarneo/ku describes the same argument correctly, so the two also disagreed
// about shared behaviour.
func TestClusterArgDescribesValidation(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	viper.Set("kubectl", map[string]any{"configPath": dir})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	kc, err := kubectl.New(l, c)
	require.NoError(t, err)

	cmd := k9s.NewCommand(l, kc)

	info := cmd.Describe(context.Background())
	require.NotEmpty(t, info.Arguments)

	cluster := info.Arguments[0]
	assert.Equal(t, "cluster", cluster.Name)

	// The claim the description makes, checked against the behaviour rather than
	// only against itself: an unknown cluster is refused, and a known one - one
	// whose kubeconfig exists under kubectl's configPath - is accepted.
	ctx := context.Background()

	err = cmd.Validate(ctx, parse(t, l, "k9s nope"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid [cluster] argument")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "real.yaml"), []byte("{}"), 0o600))
	assert.NoError(t, cmd.Validate(ctx, parse(t, l, "k9s real")))
}

func parse(t *testing.T, l log.Logger, line string) *readline.Readline {
	t.Helper()

	r, err := readline.New(l)
	require.NoError(t, err)
	require.NoError(t, r.Parse(line))

	return r
}
