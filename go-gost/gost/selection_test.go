package gost_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVerbSelection pins the arg gate both `start` and `stop` use to decide
// between the named processes and every configured one.
//
// Args at these leaves are [<verb>, <name>...], so the gate must be LenGt(1):
// it was LenGt(2), which made a single name fall through to "every configured
// process" - starting or stopping processes the caller never named.
func TestVerbSelection(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	tests := []struct {
		line  string
		names []string // nil means "every configured process"
	}{
		{"gost start", nil},
		{"gost start local", []string{"local"}},
		{"gost start local staging", []string{"local", "staging"}},
		{"gost stop", nil},
		{"gost stop local", []string{"local"}},
		{"gost stop local staging", []string{"local", "staging"}},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			t.Parallel()

			r, err := readline.New(log.NewFmt())
			require.NoError(t, err)
			require.NoError(t, r.Parse(tt.line))

			var names []string
			if r.Args().LenGt(1) {
				names = r.Args().From(1)
			}

			assert.Equal(t, tt.names, names)
		})
	}
}
