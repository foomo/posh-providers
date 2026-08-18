package testutils_test

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/foomo/posh-providers/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUndocumentedConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "a fully documented schema is clean",
			input:  `{"properties":{"path":{"type":"string","description":"Path to the root"}}}`,
			expect: nil,
		},
		{
			name:   "a missing description is reported",
			input:  `{"properties":{"path":{"type":"string"}}}`,
			expect: []string{"path"},
		},
		{
			name:   "an empty description is reported",
			input:  `{"properties":{"path":{"type":"string","description":"  "}}}`,
			expect: []string{"path"},
		},
		{
			name: "a $defs entry is walked under its own name",
			input: `{"$defs":{"Access":{"properties":{
				"hostname":{"type":"string","description":"Host"},
				"port":{"type":"integer"}}}}}`,
			expect: []string{"Access.port"},
		},
		{
			// A $ref must not be followed: the type is walked once under $defs,
			// so following it would report the same field once per use.
			name: "a $ref is not followed",
			input: `{
				"$defs":{"Access":{"properties":{"port":{"type":"integer"}}}},
				"properties":{
					"a":{"$ref":"#/$defs/Access","description":"A"},
					"b":{"$ref":"#/$defs/Access","description":"B"}}}`,
			expect: []string{"Access.port"},
		},
		{
			// A Go map reflects to additionalProperties, which is how
			// cloudflared's keyed access section is shaped.
			name: "an inline map value type is walked",
			input: `{"properties":{"access":{"type":"object","description":"Accesses",
				"additionalProperties":{"properties":{"port":{"type":"integer"}}}}}}`,
			expect: []string{"access[*].port"},
		},
		{
			name: "an inline array item type is walked",
			input: `{"properties":{"clusters":{"type":"array","description":"Clusters",
				"items":{"properties":{"name":{"type":"string"}}}}}}`,
			expect: []string{"clusters[].name"},
		},
		{
			name: "every gap is reported, sorted",
			input: `{"properties":{
				"zebra":{"type":"string"},
				"alpha":{"type":"string"}}}`,
			expect: []string{"alpha", "zebra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var schema map[string]any
			require.NoError(t, json.Unmarshal([]byte(tt.input), &schema))
			assert.Equal(t, tt.expect, testutils.UndocumentedConfig(schema))
		})
	}
}

func TestAssertDocumentedConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	documented := path.Join(dir, "documented.json")
	require.NoError(t, os.WriteFile(documented,
		[]byte(`{"properties":{"path":{"type":"string","description":"Path"}}}`), 0600))

	// A documented schema passes on the real *testing.T.
	testutils.AssertDocumentedConfig(t, documented)

	// An undocumented one fails. Verified against a throwaway T so the failure
	// does not fail this test.
	undocumented := path.Join(dir, "undocumented.json")
	require.NoError(t, os.WriteFile(undocumented,
		[]byte(`{"properties":{"path":{"type":"string"}}}`), 0600))

	fake := &testing.T{}
	testutils.AssertDocumentedConfig(fake, undocumented)
	assert.True(t, fake.Failed(), "expected an undocumented schema to fail the assertion")

	// So does a missing file, rather than passing vacuously.
	fake = &testing.T{}
	testutils.AssertDocumentedConfig(fake, path.Join(dir, "missing.json"))
	assert.True(t, fake.Failed(), "expected a missing schema to fail the assertion")
}
