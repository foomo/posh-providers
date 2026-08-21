package mjml_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/mjmlio/mjml"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPathArgumentScopesTheRun covers the [path] argument actually scoping which
// templates are compiled.
//
// files(ctx, root) built a cache key from root and then called
// files.Find(ctx, ".", "*.mjml") with a hardcoded ".", so the argument was
// validated, reported and cached-by but never applied: `mjml a` compiled every
// template in the project while logging that it was running under "a".
//
// A stub `mjml` earlier on PATH records one line per invocation.
func TestPathArgumentScopesTheRun(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	for _, project := range []string{"alpha", "beta"} {
		src := path.Join(dir, project, "src")
		require.NoError(t, os.MkdirAll(src, 0o700))
		require.NoError(t, os.WriteFile(path.Join(src, project+".mjml"),
			[]byte("<mjml></mjml>\n"), 0o600))
	}

	t.Chdir(dir)
	t.Setenv("PROJECT_ROOT", dir)

	bin := t.TempDir()
	record := path.Join(bin, "argv.txt")
	stub := "#!/bin/sh\necho \"$*\" >> " + record + "\nexit 0\n"
	require.NoError(t, os.WriteFile(path.Join(bin, "mjml"), []byte(stub), 0o700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	cmd := mjml.NewCommand(log.NewFmt(), &cache.MemoryCache{})

	r, err := readline.New(log.NewFmt())
	require.NoError(t, err)
	require.NoError(t, r.Parse("mjml alpha"))
	require.NoError(t, cmd.Execute(t.Context(), r))

	out, err := os.ReadFile(record)
	require.NoError(t, err)

	got := string(out)
	assert.Contains(t, got, "alpha", "the named path must be compiled")
	assert.NotContains(t, got, "beta",
		"a template outside the named path must not be compiled")
}
