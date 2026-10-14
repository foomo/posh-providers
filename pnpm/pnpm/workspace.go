package pnpm

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Workspace represents a pnpm-workspace.yaml file.
type Workspace struct {
	Packages []string `json:"packages" yaml:"packages"`
}

// LoadWorkspace reads filename and unmarshals it into a Workspace. It returns
// an error if the file cannot be read or does not contain valid YAML.
func LoadWorkspace(filename string) (*Workspace, error) {
	w := &Workspace{}

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	if err = yaml.Unmarshal(file, w); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	}

	return w, nil
}
