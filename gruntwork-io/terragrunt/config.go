package terragrunt

import (
	"context"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/foomo/posh/pkg/util/files"
)

// Config holds the terragrunt provider's configuration.
type Config struct {
	// Root of the terragrunt tree. Environments, sites and stacks are all
	// discovered beneath it as `<path>/envs/<env>/<site>/<stack>/terragrunt.hcl`,
	// so this single value determines the whole command surface; nothing else is
	// configured.
	Path string `json:"path" yaml:"path"`
	// Directory terragrunt downloads remote modules into, exported as
	// `TERRAGRUNT_DOWNLOAD` at shell startup. Resolved against the project root, so
	// leaving it unset points the download cache at the checkout itself.
	CachePath string `json:"cachePath" yaml:"cachePath"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// EnvNames returns the environment names discovered as subdirectories of
// EnvsPath.
func (c Config) EnvNames() []string {
	return c.dirNames(c.EnvsPath())
}

// EnvsPath returns the directory environments are discovered under.
func (c Config) EnvsPath() string {
	return path.Join(c.Path, "envs")
}

// SiteNames returns the site names discovered as subdirectories of
// SitesPath(env).
func (c Config) SiteNames(env string) []string {
	return c.dirNames(c.SitesPath(env))
}

// SitesPath returns the directory sites for env are discovered under.
func (c Config) SitesPath(env string) string {
	return path.Join(c.EnvsPath(), env)
}

// StackNames returns the stack names discovered under the env/site directory,
// sorted alphabetically, by walking for terragrunt.hcl files. It returns an
// error if the directory cannot be walked.
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

// StacksPath returns the directory stacks for the env/site pair are
// discovered under.
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
