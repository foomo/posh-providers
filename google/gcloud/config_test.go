package gcloud_test

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/google/gcloud"
	"github.com/foomo/posh-providers/pkg/testutils"
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
	require.NoError(t, reflector.AddGoComments("github.com/foomo/posh-providers/google/gcloud", "./"))
	// Config embeds onepassword.Secret, whose doc comments live in another
	// module - without this its properties reflect with no description.
	//
	// AddGoComments builds each key by dropping the last segment of the base
	// package and appending the walked directory, so reaching
	// ".../posh-providers/onepassword" needs a base one segment deeper. Passing
	// the package's own path yields ".../foomo/onepassword", which never matches.
	require.NoError(t, reflector.AddGoComments("github.com/foomo/posh-providers/onepassword/x", "../../onepassword"))
	schema := reflector.Reflect(&gcloud.Config{})
	schema.ID = "https://raw.githubusercontent.com/foomo/posh-providers/main/google/gcloud/config.schema.json"
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

	testutils.AssertDocumentedConfig(t, filename)
}
