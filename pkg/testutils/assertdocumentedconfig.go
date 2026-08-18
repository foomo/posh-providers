package testutils

import (
	"encoding/json"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

// AssertDocumentedConfig asserts that every property in a generated
// config.schema.json carries a description.
//
// The schema is what an agent reads to understand a provider's config key - it
// is referenced from a project's own posh config `$schema`, so it reaches every
// consumer, and it is generated from the Config struct rather than
// hand-maintained. A property without a description is therefore a field an
// agent can see but cannot interpret, and the fix is a doc comment on the struct
// field: the reflector lifts those into descriptions via AddGoComments.
//
// It asserts on the generated artifact rather than the struct because that is
// what consumers read, and because it catches every cause of a missing
// description - an absent doc comment, but also a comment the reflector did not
// pick up.
//
// Call it from a provider's config test, after the schema has been written:
//
//	func TestConfigDocumented(t *testing.T) {
//		testutils.AssertDocumentedConfig(t, "config.schema.json")
//	}
func AssertDocumentedConfig(t *testing.T, filename string) {
	t.Helper()

	body, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("failed to read %s: %v", filename, err)

		return
	}

	var schema map[string]any
	if err := json.Unmarshal(body, &schema); err != nil {
		t.Errorf("failed to parse %s: %v", filename, err)

		return
	}

	if values := UndocumentedConfig(schema); len(values) > 0 {
		t.Errorf("%s has undocumented properties - add a doc comment to the matching struct field:\n  %s",
			filename, strings.Join(values, "\n  "))
	}
}

// UndocumentedConfig walks a parsed JSON schema and reports every property
// missing a description, as a sorted list of dotted paths. It returns nil when
// the schema is fully documented.
//
// Both the top level schema and every $defs entry are walked, since the
// reflector hoists named struct types into $defs and leaves only a $ref behind.
// A $ref is not followed: the type it names is walked once under its own $defs
// entry, so following it would report the same field once per use.
func UndocumentedConfig(schema map[string]any) []string {
	var ret []string

	for name, def := range mapOf(schema, "$defs") {
		if value, ok := def.(map[string]any); ok {
			walkUndocumented(value, name, &ret)
		}
	}

	walkUndocumented(schema, "", &ret)

	sort.Strings(ret)

	return slices.Compact(ret)
}

// walkUndocumented reports undocumented properties of a schema node,
// accumulating the dotted path as it descends.
func walkUndocumented(node map[string]any, path string, ret *[]string) {
	for name, value := range mapOf(node, "properties") {
		prop, ok := value.(map[string]any)
		if !ok {
			continue
		}

		child := name
		if path != "" {
			child = path + "." + name
		}

		if description, _ := prop["description"].(string); strings.TrimSpace(description) == "" {
			*ret = append(*ret, child)
		}

		walkProperty(prop, child, ret)
	}
}

// walkProperty descends into a property's nested shapes: an inline object, the
// items of an array, or the value type of a map. Only inline definitions are
// followed - a $ref names a type walked under its own $defs entry.
func walkProperty(prop map[string]any, path string, ret *[]string) {
	walkUndocumented(prop, path, ret)

	if value, ok := prop["items"].(map[string]any); ok {
		walkProperty(value, path+"[]", ret)
	}

	// A Go map reflects to additionalProperties holding the value schema, which
	// is how a keyed config section like cloudflared's access map is shaped.
	if value, ok := prop["additionalProperties"].(map[string]any); ok {
		walkProperty(value, path+"[*]", ret)
	}
}

// mapOf returns the named child object of a schema node, or nil.
func mapOf(node map[string]any, key string) map[string]any {
	ret, _ := node[key].(map[string]any)

	return ret
}
