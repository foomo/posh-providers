package slack_test

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/slack-go/slack"
	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	cwd, err := os.Getwd()
	require.NoError(t, err)

	reflector := new(jsonschema.Reflector)
	reflector.RequiredFromJSONSchemaTags = true
	require.NoError(t, reflector.AddGoComments("github.com/foomo/posh-providers/slack-go/slack", "./"))
	// Config embeds onepassword.Secret, documented in another module.
	// AddGoComments drops the base package's last segment before appending
	// the walked dir, so the base needs one extra segment to resolve.
	require.NoError(t, reflector.AddGoComments("github.com/foomo/posh-providers/onepassword/x", "../../onepassword"))
	schema := reflector.Reflect(&slack.Config{})
	schema.ID = "https://github.com/foomo/posh-providers/slack-go/slack"
	actual, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)

	filename := path.Join(cwd, "config.schema.json")

	expected, err := os.ReadFile(filename)
	if !errors.Is(err, os.ErrNotExist) {
		require.NoError(t, err)
	}

	if !assert.Equal(t, string(expected), string(actual)) {
		require.NoError(t, os.WriteFile(filename, actual, 0600))
	}
}
