package bruno

import (
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
)

// Config holds the location of the Bruno collection the command operates on.
type Config struct {
	// Path to the Bruno collection; environment variables are expanded
	Path string `json:"path" yaml:"path" mapstructure:"path"`
}

// Filename returns the collection path with environment variables expanded.
func (c Config) Filename() string {
	return os.ExpandEnv(c.Path)
}

// Environments returns the names of the .bru environment files found in the
// collection's "environments" directory.
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

// Requests returns the sorted paths of every .bru request file in the
// collection, excluding the "environments" directory.
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
