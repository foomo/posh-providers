package hygen_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/jondot/hygen"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommandTreeShape pins the two things the rendered usage block says about
// how this command is invoked.
//
// Both were wrong before: the leaf was named "template", which reads as a literal
// subcommand an agent should type even though the node is declared with Values
// and matches any template name; and the root carried a `path` argument that no
// input can reach, which reads as a second way to invoke the command.
//
// The catalog publishes these names, so they are the only thing an agent has to
// go on - a misleading one costs it a failed run.
func TestCommandTreeShape(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	templates := filepath.Join(dir, "scaffold")
	require.NoError(t, os.MkdirAll(filepath.Join(templates, "mycomponent"), 0o755))

	viper.Set("hygen", map[string]any{"templatePath": templates})
	t.Cleanup(viper.Reset)

	cmd, err := hygen.NewCommand(log.NewFmt(), &cache.MemoryCache{})
	require.NoError(t, err)

	info := cmd.Describe(context.Background())

	// The root takes no arguments of its own: its only child is a Values node,
	// so position 0 always resolves there.
	assert.Empty(t, info.Arguments, "the root must declare no unreachable argument")

	require.Len(t, info.Subcommands, 1)
	leaf := info.Subcommands[0]

	// describe() angle-brackets a Values node, so the placeholder reads as one.
	// The name inside it still has to say what the value is: "<template>" named
	// the node after the concept, which is what made it look like a literal
	// segment an agent should type.
	assert.True(t, leaf.Dynamic, "the leaf resolves its name at runtime")
	assert.Equal(t, "hygen <template-name>", leaf.FullPath)
}

// TestCompletionResolvesTemplateNames is the behavioural half: it shows why the
// root argument was unreachable, rather than trusting the tree shape alone.
func TestCompletionResolvesTemplateNames(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	dir := t.TempDir()
	templates := filepath.Join(dir, "scaffold")
	require.NoError(t, os.MkdirAll(filepath.Join(templates, "mycomponent"), 0o755))

	viper.Set("hygen", map[string]any{"templatePath": templates})
	t.Cleanup(viper.Reset)

	l := log.NewFmt()

	cmd, err := hygen.NewCommand(l, &cache.MemoryCache{})
	require.NoError(t, err)

	r := parse(t, l, "hygen ")

	var got []string
	for _, s := range cmd.Complete(context.Background(), r) {
		got = append(got, s.Text)
	}

	// Template names, not paths from the root argument's own Suggest.
	assert.Equal(t, []string{"mycomponent"}, got)
}

func parse(t *testing.T, l log.Logger, line string) *readline.Readline {
	t.Helper()

	r, err := readline.New(l)
	require.NoError(t, err)
	require.NoError(t, r.Parse(line))

	return r
}
