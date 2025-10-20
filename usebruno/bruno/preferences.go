package bruno

import (
	"os"
	"slices"

	"github.com/goccy/go-json"
	"github.com/mitchellh/mapstructure"
)

type Preferences struct {
	path string
	data map[string]any
}

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

func (p *Preferences) Save(path string) error {
	data, err := json.MarshalIndent(p.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
