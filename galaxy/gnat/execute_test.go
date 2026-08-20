package gnat_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/galaxy/gnat"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecuteConfigDir covers how configDir reaches the gnat process.
//
// execute used to append the XDG_CONFIG_HOME=... string with .Args(), which
// hands it to gnat as a trailing positional argument and leaves the variable
// unset - so gnat read and wrote its config under the user's real
// $XDG_CONFIG_HOME, exactly what the option exists to prevent. shell.Args
// appends to argv; shell.Env sets cmd.Env. The sibling galaxy/tempo uses .Env()
// and was already correct.
//
// A stub `gnat` earlier on PATH records what the real call site produced, so
// this asserts the provider's own behaviour rather than shell's semantics.
//
// Not parallel: NewCommand reads config through viper, which is a singleton,
// and this mutates PATH.
func TestExecuteConfigDir(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	const configDir = "/proj/.posh/config/gnat"

	bin := t.TempDir()
	record := path.Join(bin, "invocation.txt")
	stub := "#!/bin/sh\n{ echo \"ARGV=$*\"; echo \"ENV=$XDG_CONFIG_HOME\"; } > " + record + "\n"
	require.NoError(t, os.WriteFile(path.Join(bin, "gnat"), []byte(stub), 0o700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	viper.Set("gnat", map[string]any{
		"configDir": configDir,
		"profiles":  map[string]any{"local": map[string]any{"url": "nats://localhost:4222"}},
	})
	t.Cleanup(viper.Reset)

	cmd, err := gnat.NewCommand(log.NewFmt())
	require.NoError(t, err)

	r, err := readline.New(log.NewFmt())
	require.NoError(t, err)
	require.NoError(t, r.Parse("gnat local"))
	require.NoError(t, cmd.Execute(t.Context(), r))

	out, err := os.ReadFile(record)
	require.NoError(t, err)

	got := string(out)
	assert.Contains(t, got, "ENV="+configDir,
		"configDir must be set as XDG_CONFIG_HOME in the process environment")
	assert.Contains(t, got, "ARGV=-url nats://localhost:4222\n",
		"argv must carry only the profile URL, not the environment assignment")
	assert.NotContains(t, got, "ARGV=-url nats://localhost:4222 XDG_CONFIG_HOME",
		"the variable must not be passed to gnat as a positional argument")
}
