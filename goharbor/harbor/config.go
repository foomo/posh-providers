package harbor

import (
	"strings"
)

type Config struct {
	URL     string `json:"url" yaml:"url"`
	AuthURL string `json:"authUrl" yaml:"authUrl"`
	Project string `json:"project" yaml:"project"`
}

func (c Config) DockerRegistry() string {
	return strings.TrimPrefix(c.URL, "https://")
}
