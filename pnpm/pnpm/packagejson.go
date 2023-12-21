package pnpm

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

// PackageJSON represents NodeJS package.json
type PackageJSON struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords"`
	Homepage    string            `json:"homepage"`
	License     string            `json:"license"`
	Files       []string          `json:"files"`
	Main        string            `json:"main"`
	Scripts     map[string]string `json:"scripts"`
	OS          []string          `json:"os"`
	CPU         []string          `json:"cpu"`
	Private     bool              `json:"private"`
}

func LoadPackageJSON(filename string) (*PackageJSON, error) {
	w := &PackageJSON{}

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	if err = json.Unmarshal(file, w); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	}

	return w, nil
}
