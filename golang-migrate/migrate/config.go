package migrate

import (
	"github.com/samber/lo"
)

type Config struct {
	// SourcesMap maps a source name to a golang-migrate source URL, e.g.
	// `file://migrations`. The names are the values the `source` argument accepts.
	SourcesMap map[string]string `json:"sources" yaml:"sources" mapstructure:"sources"`
	// DatabasesMap maps a database name to a golang-migrate database URL, e.g.
	// `pgx5://user:password@localhost:5432/db?sslmode=disable`. The names are the
	// values the `database` argument accepts. A value may contain 1Password
	// secret references, which are resolved at execution time.
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
