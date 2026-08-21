package rclone_test

import (
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/rclone/rclone"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRemoteSuggestions covers what remote completion offers.
//
// The suggester did strings.Split(strings.Trim(out, "\n"), "\n"), and Split on an
// empty string returns [""] rather than an empty slice - so with no remotes
// configured, completion showed a single blank entry, reading as "one remote
// exists" rather than "none". The exec error was discarded too, making a missing
// binary indistinguishable from an empty config.
func TestRemoteSuggestions(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{"no remotes configured", "", nil},
		{"trailing newline only", "\n", nil},
		{"one remote", "cloudflare:\n", []string{"cloudflare:"}},
		{"several remotes", "cloudflare:\ns3:\n", []string{"cloudflare:", "s3:"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bin := t.TempDir()
			stub := "#!/bin/sh\nprintf '%s' '" + tt.output + "'\nexit 0\n"
			require.NoError(t, os.WriteFile(path.Join(bin, "rclone"), []byte(stub), 0o700))
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

			viper.Set("rclone", map[string]any{"path": path.Join(t.TempDir(), "rclone.conf")})
			t.Cleanup(viper.Reset)

			cmd, err := rclone.NewCommand(log.NewFmt(), &cache.MemoryCache{})
			require.NoError(t, err)

			r, err := readline.New(log.NewFmt())
			require.NoError(t, err)
			require.NoError(t, r.Parse("rclone ls "))

			var got []string
			for _, s := range cmd.Complete(t.Context(), r) {
				got = append(got, s.Text)
			}

			assert.Equal(t, tt.want, got)

			for _, s := range got {
				assert.NotEmpty(t, s, "a blank suggestion reads as an existing remote")
			}
		})
	}
}
