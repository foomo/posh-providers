package pnpm

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Workspace struct {
	Packages []string `json:"packages" yaml:"packages"`
}

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
