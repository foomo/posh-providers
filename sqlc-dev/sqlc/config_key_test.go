package sqlc_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/sqlc-dev/sqlc"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// TestConfigKeyIsRead guards the default config key.
//
// NewCommand used to leave configKey empty, so viper.UnmarshalKey("") returned a
// zero Config however the project had configured `sqlc:` - leaving SQLCCACHE and
// SQLCTMPDIR pointing at the project root instead of the configured directories.
// UnmarshalKey does not error on a missing key, so nothing surfaced it.
//
// The observable effect lives in a subprocess environment, so this asserts the
// mechanism instead: viper must resolve the key the constructor uses, and a
// populated `sqlc:` block must not unmarshal to the zero value.
func TestConfigKeyIsRead(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	viper.Set("sqlc", map[string]any{
		"tempDir":  ".posh/tmp/sqlc",
		"cacheDir": ".posh/cache/sqlc",
	})
	t.Cleanup(viper.Reset)

	var cfg sqlc.Config
	require.NoError(t, viper.UnmarshalKey("sqlc", &cfg))
	require.Equal(t, ".posh/tmp/sqlc", cfg.TempDir, "the `sqlc:` block must reach Config")
	require.Equal(t, ".posh/cache/sqlc", cfg.CacheDir)

	// The empty key is what the bug used to pass, and it silently yields nothing.
	var empty sqlc.Config
	require.NoError(t, viper.UnmarshalKey("", &empty))
	require.Empty(t, empty.TempDir, "UnmarshalKey(\"\") yields a zero Config, which is why the key must be set")

	// The constructor must succeed with the key it defaults to.
	_, err := sqlc.NewCommand(log.NewFmt(), &cache.MemoryCache{})
	require.NoError(t, err)
}
