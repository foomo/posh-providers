package migrate

import (
	"github.com/samber/lo"
)

type Config struct {
	SourcesMap   map[string]string `json:"sources" yaml:"sources" mapstructure:"sources"`
	DatabasesMap map[string]string `json:"databases" yaml:"databases" mapstructure:"databases"`
}

func (c Config) Sources() []string {
	return lo.Keys(c.SourcesMap)
}

func (c Config) Source(name string) string {
	if value, ok := c.SourcesMap[name]; ok {
		return value
	}
	return ""
}

func (c Config) Databases() []string {
	return lo.Keys(c.DatabasesMap)
}

func (c Config) Database(name string) string {
	if value, ok := c.DatabasesMap[name]; ok {
		return value
	}
	return ""
}
