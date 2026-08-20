package gokazi_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStopSelection pins the arg gate `stop` uses to decide between the named
// tasks and every task in the registry.
//
// Args at this leaf are ["stop", <name>...], so the gate must be LenGt(1): it
// was LenGt(2), which made a single name fall through to "every task" and stop
// port forwards and tunnels the caller never named.
func TestStopSelection(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	tests := []struct {
		line  string
		names []string // nil means "every task in the registry"
	}{
		{"gokazi stop", nil},
		{"gokazi stop taskA", []string{"taskA"}},
		{"gokazi stop taskA taskB", []string{"taskA", "taskB"}},
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
