package task

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// Path to the directory containing tasks
	Path string `json:"path" yaml:"path"`
	// Task configurations
	Tasks map[string]Task `json:"tasks" yaml:"tasks"`
}

func (c Config) Names() []string {
	var ret []string

	for k, v := range c.Tasks {
		if !v.Hidden {
			ret = append(ret, k)
		}
	}

	// Read YAML files from the path
	if c.Path != "" {
		if entries, err := os.ReadDir(c.Path); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					name := entry.Name()
					if strings.HasSuffix(name, ".yaml") {
						// Remove the extension
						nameWithoutExt := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
						ret = append(ret, nameWithoutExt)
					}
				}
			}
		}
	}

	sort.Strings(ret)

	return ret
}

func (c Config) AllTasks() (map[string]Task, error) {
	ret := c.Tasks
	if c.Path != "" {
		if entries, err := os.ReadDir(c.Path); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					name := entry.Name()
					if strings.HasSuffix(name, ".yaml") {
						data, err := os.ReadFile(filepath.Join(c.Path, name))
						if err != nil {
							return nil, errors.Wrap(err, "failed to read task file "+name)
						}

						var task Task
						if err := yaml.Unmarshal(data, &task); err != nil {
							return nil, errors.Wrap(err, "failed to unmarshal task file "+name)
						}

						ret[strings.TrimSuffix(name, ".yaml")] = task
					}
				}
			}
		}
	}

	return ret, nil
}
