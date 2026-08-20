package kubectl_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/kubernetes/kubectl"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/readline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProfile covers reading --profile from raw, unparsed flag tokens.
//
// Validate runs before Execute, and flag sets are only built inside Execute, so
// r.FlagSets() is nil at validation time - the raw tokens in r.Flags() are the
// only place the profile can be read from.
func TestProfile(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	tests := map[string]string{
		"helm install mycluster":                           "",
		"helm install mycluster --profile staging":         "staging",
		"helm install mycluster --profile=staging":         "staging",
		"helm install mycluster --debug --profile staging": "staging",
		"helm install mycluster --profile staging --debug": "staging",
		"helm install mycluster --profile":                 "",
		"helm install mycluster --namespace profile":       "",
	}

	for line, want := range tests {
		t.Run(line, func(t *testing.T) {
			t.Parallel()

			r, err := readline.New(log.NewFmt())
			require.NoError(t, err)
			require.NoError(t, r.Parse(line))

			assert.Equal(t, want, kubectl.Profile(r.Flags()))
		})
	}
}
