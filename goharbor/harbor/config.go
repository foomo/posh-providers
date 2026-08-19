package harbor

import (
	"strings"
)

type Config struct {
	// Harbor base URL including scheme, e.g. `https://harbor.example.org`. Passed
	// to `docker login` as-is; the registry host is derived from it by stripping a
	// `https://` prefix, so an `http://` URL or a trailing slash is not handled.
	URL string `json:"url" yaml:"url"`
	// Login page opened in a browser by `auth`, typically the OIDC entry point
	// (`<url>/c/oidc/login`). Opened by `docker` too, so the CLI secret can be
	// copied from the profile page.
	AuthURL string `json:"authUrl" yaml:"authUrl"`
	// Harbor project used to build the throwaway image reference the auth check
	// pulls. Not used by `auth` or `docker`.
	Project string `json:"project" yaml:"project"`
}

func (c Config) DockerRegistry() string {
	return strings.TrimPrefix(c.URL, "https://")
}
