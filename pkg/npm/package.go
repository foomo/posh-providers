package npm

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

// Package represents NodeJS package.json
type Package struct {
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	Version          string            `json:"version"`
	Description      string            `json:"description"`
	Keywords         []string          `json:"keywords"`
	Homepage         string            `json:"homepage"`
	License          string            `json:"license"`
	Files            []string          `json:"files"`
	Main             string            `json:"main"`
	PackageManager   string            `json:"packageManager"`
	Scripts          map[string]string `json:"scripts"`
	Exports          []string          `json:"exports"`
	OS               []string          `json:"os"`
	CPU              []string          `json:"cpu"`
	Private          bool              `json:"private"`
	Dependencies     map[string]string `json:"dependencies"`
	DevDependencies  map[string]string `json:"devDependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
	Overrides        map[string]string `json:"overrides"`
	Resolutions      map[string]string `json:"resolutions"`
	Workspaces       Workspaces        `json:"workspaces"`
}

func LoadPackageJSON(filename string) (*Package, error) {
	w := &Package{}

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	if err = json.Unmarshal(file, w); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	}

	return w, nil
}
