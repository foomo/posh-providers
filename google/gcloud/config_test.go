package gcloud_test

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/posh-providers/google/gcloud"
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
	require.NoError(t, reflector.AddGoComments("github.com/foomo/posh-providers/google/gcloud", "./"))
	schema := reflector.Reflect(&gcloud.Config{})
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
