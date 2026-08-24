package gotsrpc_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/foomo/gotsrpc"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlagRewrite covers converting posh's double-dash flags to the single-dash
// long flags gotsrpc expects.
//
// The rewrite used to be strings.ReplaceAll(flag, "--", "-"), a blind
// whole-string replacement - so a flag whose *value* contained "--" was
// corrupted too, turning --out=a--b into -out=a-b. Only the prefix should change.
//
// A stub `gotsrpc` earlier on PATH records the argv it was called with.
func TestFlagRewrite(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(path.Join(dir, "gotsrpc.yml"), []byte("targets: {}\n"), 0o600))
	t.Chdir(dir)
	t.Setenv("PROJECT_ROOT", dir)

	bin := t.TempDir()
	record := path.Join(bin, "argv.txt")
	stub := "#!/bin/sh\necho \"$*\" >> " + record + "\nexit 0\n"
	require.NoError(t, os.WriteFile(path.Join(bin, "gotsrpc"), []byte(stub), 0o700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	cmd := gotsrpc.NewCommand(log.NewFmt(), &cache.MemoryCache{})

	r, err := readline.New(log.NewFmt())
	require.NoError(t, err)
	// --out carries a value containing a double dash, which must survive intact.
	require.NoError(t, r.Parse("gotsrpc . --out=a--b"))
	_ = cmd.Execute(t.Context(), r)

	out, err := os.ReadFile(record)
	require.NoError(t, err)

	got := string(out)
	assert.Contains(t, got, "-out=a--b", "only the flag prefix may be rewritten")
	assert.NotContains(t, got, "-out=a-b", "the flag's value must not be rewritten")
}
