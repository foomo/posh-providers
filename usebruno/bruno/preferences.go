package bruno

import (
	"os"
	"slices"

	"github.com/goccy/go-json"
	"github.com/mitchellh/mapstructure"
)

// Preferences holds the Bruno desktop app's preferences.json contents, loaded
// from and saved back to disk.
type Preferences struct {
	path string
	data map[string]any
}

// NewPreferences reads and parses the preferences.json file at path. It
// returns an error if the file cannot be read or does not contain valid JSON.
func NewPreferences(path string) (*Preferences, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	return &Preferences{
		path: path,
		data: data,
	}, nil
}

// AddLastOpenedCollection appends path to the "lastOpenedCollections" list if
// it is not already present. It is a no-op when the existing value is not a
// list.
func (p *Preferences) AddLastOpenedCollection(path string) error {
	if _, ok := p.data["lastOpenedCollections"]; !ok {
		p.data["lastOpenedCollections"] = make(map[string][]any)
	}

	if lastOpenedCollections, ok := p.data["lastOpenedCollections"].([]any); ok {
		var lastOpenedCollectionsStrings []string
		if err := mapstructure.Decode(lastOpenedCollections, &lastOpenedCollectionsStrings); err != nil {
			return err
		}

		if !slices.Contains(lastOpenedCollectionsStrings, path) {
			lastOpenedCollections = append(lastOpenedCollections, path)
			p.data["lastOpenedCollections"] = lastOpenedCollections
		}
	}

	return nil
}

// Save writes the preferences as indented JSON to path with 0600 permissions.
func (p *Preferences) Save(path string) error {
	data, err := json.MarshalIndent(p.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
