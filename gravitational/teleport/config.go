package teleport

import (
	"fmt"
	"os"
	"strings"
)

type (
	Config struct {
		Path     string            `json:"path" yaml:"path"`
		Labels   map[string]string `json:"labels" yaml:"labels"`
		Hostname string            `json:"hostname" yaml:"hostname"`
		Database Database          `json:"database" yaml:"database"`
	}
	Database struct {
		User string `json:"user" yaml:"user"`
	}
)

func (c Config) Query() string {
	if len(c.Labels) == 0 {
		return ""
	}
	var ret []string
	for k, v := range c.Labels {
		ret = append(ret, fmt.Sprintf("labels[\"%s\"] == \"%s\"", k, v))
	}
	return strings.Join(ret, " && ")
}

func (c Database) EnvUser() string {
	if value := os.Getenv("TELEPORT_DATABASE_USER"); value != "" {
		return value
	}
	return c.User
}
