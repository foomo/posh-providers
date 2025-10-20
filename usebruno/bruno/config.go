package bruno

import (
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
)

type Config struct {
	Path string `json:"path" yaml:"path" mapstructure:"path"`
}

func (c Config) Filename() string {
	return os.ExpandEnv(c.Path)
}

func (c Config) Environments() []string {
	entries, err := fs.Glob(os.DirFS(path.Join(c.Filename(), "environments")), "*.bru")
	if err != nil {
		return nil
	}

	var ret []string
	for _, entry := range entries {
		ret = append(ret, strings.TrimSuffix(entry, ".bru"))
	}

	return ret
}

func (c Config) Requests() []string {
	var (
		ret   []string
		files []string
	)

	if value, err := fs.Glob(os.DirFS(c.Filename()), "*.bru"); err == nil {
		files = append(files, value...)
	}

	if value, err := fs.Glob(os.DirFS(c.Filename()), "**/*.bru"); err == nil {
		files = append(files, value...)
	}

	slices.Sort(files)

	for _, entry := range files {
		if !strings.HasPrefix(entry, "environments") {
			ret = append(ret, entry)
		}
	}

	return ret
}
