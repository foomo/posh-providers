package mkcert_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/filosottile/mkcert"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForwardedArgs covers what each verb passes to the mkcert binary.
//
// install, caroot and uninstall each forwarded r.Args() - which at a first-level
// leaf is the node's own name - so `mkcert install` ran `mkcert -install
// install`. Upstream's -install only returns early when len(args) == 0, and
// "install" matches its hostname regex, so it installed the CA and then issued a
// real certificate for a host named "install", writing install.pem and
// install-key.pem. These verbs set no Dir(), so that was the project root, and
// upstream does no existence check before overwriting.
//
// caroot and uninstall were unaffected only because upstream returns before
// reading positionals - the same wrong line, inert on two verbs and damaging on
// the third.
//
// A stub `mkcert` earlier on PATH records the argv it was called with.
func TestForwardedArgs(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	certPath := t.TempDir()
	viper.Set("mkcert", map[string]any{"certificatePath": certPath})
	t.Cleanup(viper.Reset)

	bin := t.TempDir()
	record := path.Join(bin, "argv.txt")
	stub := "#!/bin/sh\necho \"$*\" >> " + record + "\n"
	require.NoError(t, os.WriteFile(path.Join(bin, "mkcert"), []byte(stub), 0o700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	cmd, err := mkcert.NewCommand(log.NewFmt())
	require.NoError(t, err)

	tests := map[string]string{
		"mkcert install":         "-install",
		"mkcert caroot":          "-CAROOT",
		"mkcert uninstall":       "-uninstall",
		"mkcert create foo.test": "foo.test",
	}

	for line, want := range tests {
		t.Run(line, func(t *testing.T) {
			require.NoError(t, os.WriteFile(record, nil, 0o600))

			r, err := readline.New(log.NewFmt())
			require.NoError(t, err)
			require.NoError(t, r.Parse(line))
			require.NoError(t, cmd.Execute(t.Context(), r))

			out, err := os.ReadFile(record)
			require.NoError(t, err)

			assert.Equal(t, want+"\n", string(out),
				"the verb must not forward its own node name to mkcert")
		})
	}
}
