package lint_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/arbitrary/lint"
	"github.com/foomo/posh/pkg/command"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubLinter satisfies both command.Command (to sit in the registry, which is
// where linters are discovered from by type assertion) and lint.Linter.
type stubLinter struct {
	name   string
	linted atomic.Bool
}

func (s *stubLinter) Name() string        { return s.name }
func (s *stubLinter) Description() string { return "stub " + s.name }

func (s *stubLinter) Execute(ctx context.Context, r *readline.Readline) error { return nil }

func (s *stubLinter) Lint(ctx context.Context, fix bool) error {
	s.linted.Store(true)

	return nil
}

// warnLogger records Warn calls so the per-name warning can be asserted.
// log.Fmt writes to stdout with no injectable writer, so it is embedded and only
// the two methods that matter are overridden - Named because the constructor
// calls it, and it must keep returning this recorder.
type warnLogger struct {
	*log.Fmt

	mu    sync.Mutex
	warns []string
}

func (l *warnLogger) Named(string) log.Logger { return l }

func (l *warnLogger) Warn(a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	parts := make([]string, 0, len(a))
	for _, v := range a {
		parts = append(parts, fmt.Sprint(v))
	}

	l.warns = append(l.warns, strings.Join(parts, " "))
}

func (l *warnLogger) Warnings() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return slices.Clone(l.warns)
}

// TestUnknownLinterNames covers naming a linter that is not registered.
//
// The filter kept the registered linters whose Name() appeared in the args and
// errored only when the *result* was empty - so `lint go bogus` ran `go` and
// never mentioned `bogus`. Each unmatched name is now warned about individually
// while the linters that did match still run, so a partial invocation is not
// blocked but neither is it silent.
//
// Both halves are asserted: the warning emitted per unmatched name, and which
// linters actually ran.
func TestUnknownLinterNames(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	tests := []struct {
		line       string
		wantErr    string
		wantLinted []string
		wantWarns  []string
	}{
		{"lint", "", []string{"go", "helm"}, nil},
		{"lint go", "", []string{"go"}, nil},
		{"lint go helm", "", []string{"go", "helm"}, nil},
		// The case that used to pass silently: `go` still runs, and `bogus` is
		// warned about rather than ignored.
		{"lint go bogus", "", []string{"go"}, []string{"unknown linter: bogus"}},
		// Nothing matched at all, so there is no work to do and it stays an error
		// rather than a silent success.
		{"lint bogus", "unknown linter: [bogus]", nil, []string{"unknown linter: bogus"}},
		{"lint bogus other", "unknown linter: [bogus other]", nil, []string{"unknown linter: bogus", "unknown linter: other"}},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			gol, helm := &stubLinter{name: "go"}, &stubLinter{name: "helm"}

			commands := command.Commands{}
			commands.Add(gol, helm)

			l := &warnLogger{Fmt: log.NewFmt()}
			cmd := lint.NewCommand(l, commands)

			r, err := readline.New(log.NewFmt())
			require.NoError(t, err)
			require.NoError(t, r.Parse(tt.line))

			err = cmd.Execute(t.Context(), r)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErr)
			}

			assert.Equal(t, tt.wantWarns, nilIfEmpty(l.Warnings()),
				"every unmatched name must be warned about individually")

			// The linters that matched must run; the ones that did not must not.
			for _, lt := range []*stubLinter{gol, helm} {
				want := slices.Contains(tt.wantLinted, lt.name)
				assert.Equal(t, want, lt.linted.Load(), "linter %q ran unexpectedly or was skipped", lt.name)
			}
		})
	}
}

func nilIfEmpty(v []string) []string {
	if len(v) == 0 {
		return nil
	}

	return v
}
