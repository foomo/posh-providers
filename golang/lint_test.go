package golang_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLintForwarding pins the flags `lint` hands to golangci-lint.
//
// It used to forward both fs.Visited().Args() (re-rendered from pflag) and
// r.Flags() (every raw token typed), which had two consequences:
//
//   - `--parallel` is declared on the *internal* flag set precisely so it stays
//     inside posh, but r.Flags() does not respect that split - so it reached
//     golangci-lint, which has no such flag and exits `unknown flag: --parallel`
//     (confirmed against golangci-lint 2.12.2). The one node declaring
//     --parallel was the one node where it could not be used.
//   - every default flag arrived twice, once per source.
//
// Only the parsed set is forwarded now, matching every sibling node in this
// file - none of which forwards r.Flags() at all.
func TestLintForwarding(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	tests := []struct {
		line string
		want []string
	}{
		{"go lint", nil},
		// --parallel is internal: it must become --allow-parallel-runners and
		// must NOT be forwarded verbatim.
		{"go lint --parallel 4", []string{"--allow-parallel-runners"}},
		{"go lint --timeout 5m", []string{"--timeout", "5m0s"}},
		{"go lint --fix", []string{"--fix"}},
		{"go lint --parallel 2 --fix", []string{"--allow-parallel-runners", "--fix"}},
		// anything after -- is forwarded untouched
		{"go lint -- --verbose", []string{"--", "--verbose"}},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			t.Parallel()

			r, err := readline.New(log.NewFmt())
			require.NoError(t, err)
			require.NoError(t, r.Parse(tt.line))

			// Mirrors the lint node's flag set.
			fs := readline.NewFlagSets()
			fs.Default().Duration("timeout", 0, "Timeout for total work")
			fs.Default().Bool("fast-only", false, "")
			fs.Default().Bool("new", false, "")
			fs.Default().Bool("fix", false, "")
			fs.Default().String("out-format", "", "")
			fs.Default().Int("concurrency", 0, "")
			fs.Internal().Int("parallel", 0, "Number of parallel processes")
			r.SetFlagSets(fs)
			require.NoError(t, r.ParseFlagSets())

			// Mirrors lint()'s argument assembly.
			var args []string
			if value, _ := fs.Internal().GetInt("parallel"); value != 0 {
				args = append(args, "--allow-parallel-runners")
			}

			args = append(args, fs.Default().Visited().Args()...)
			args = append(args, r.AdditionalArgs()...)

			assert.Equal(t, tt.want, args)
			assert.NotContains(t, args, "--parallel", "the internal --parallel flag must not reach golangci-lint")
		})
	}
}
