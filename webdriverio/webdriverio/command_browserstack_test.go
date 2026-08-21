package webdriverio_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh-providers/webdriverio/webdriverio"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBrowserstackModeWithoutSecret covers selecting the browserstack mode
// without configuring the secret it needs.
//
// "browserstack" is a magic mode name rather than a declared one, and execute
// dereferenced *c.cfg.BrowserStack unconditionally in that branch - so a project
// defining a mode by that name without a browserStack: key crashed instead of
// erroring. --ci short-circuits the branch, so whether it panicked depended on
// which flags were passed.
//
// Not parallel: both constructors read config through viper, which is a
// singleton.
func TestBrowserstackModeWithoutSecret(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	viper.Set("webdriverio", map[string]any{
		"modes": map[string]any{"browserstack": map[string]any{}},
		"sites": map[string]any{
			"shop": map[string]any{"dev": map[string]any{"host": "shop.example.com"}},
		},
		// browserStack deliberately unset
	})
	t.Cleanup(viper.Reset)

	l, c := log.NewFmt(), &cache.MemoryCache{}

	op, err := onepassword.New(l, c)
	require.NoError(t, err)

	cmd, err := webdriverio.NewCommand(l, c, op)
	require.NoError(t, err)

	r, err := readline.New(l)
	require.NoError(t, err)
	require.NoError(t, r.Parse("wdio browserstack shop dev"))

	var execErr error

	assert.NotPanics(t, func() {
		execErr = cmd.Execute(t.Context(), r)
	}, "an unset browserStack config must not panic")

	require.Error(t, execErr, "it must report the missing config instead")
	assert.Contains(t, execErr.Error(), "browserStack",
		"the error should name the config key that is missing")
}
