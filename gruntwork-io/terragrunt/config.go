package terragrunt

import (
	"context"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/foomo/posh/pkg/util/files"
)

type Config struct {
	Path      string `json:"path" yaml:"path"`
	CachePath string `json:"cachePath" yaml:"cachePath"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c Config) EnvNames() []string {
	return c.dirNames(c.EnvsPath())
}

func (c Config) EnvsPath() string {
	return path.Join(c.Path, "envs")
}

func (c Config) SiteNames(env string) []string {
	return c.dirNames(c.SitesPath(env))
}

func (c Config) SitesPath(env string) string {
	return path.Join(c.EnvsPath(), env)
}

func (c Config) StackNames(ctx context.Context, env, site string) ([]string, error) {
	var ret []string

	root := c.StacksPath(env, site)

	out, err := files.Find(ctx, root, "terragrunt.hcl", files.FindWithIsFile(true))
	if err != nil {
		return nil, err
	}

	for _, v := range out {
		v = strings.TrimPrefix(v, root)

		v = strings.TrimSuffix(v, "/terragrunt.hcl")
		if v = strings.TrimPrefix(v, "/"); strings.Contains(v, "/") {
			ret = append(ret, v)
		}
	}

	slices.Sort(ret)

	return ret, err
}

func (c Config) StacksPath(env, site string) string {
	return path.Join(c.SitesPath(env), site)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c Config) dirNames(path string) []string {
	var ret []string

	if files, err := os.ReadDir(path); err == nil {
		for _, value := range files {
			if value.IsDir() && !strings.HasPrefix(value.Name(), ".") && !strings.HasPrefix(value.Name(), "_") {
				ret = append(ret, value.Name())
			}
		}
	}

	return ret
}
