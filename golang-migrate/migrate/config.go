package migrate

import (
	"github.com/samber/lo"
)

// Config is the migrate provider configuration: named migration sources and
// named database connection URLs.
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

// Sources returns the configured migration source names, in no particular
// order.
func (c Config) Sources() []string {
	return lo.Keys(c.SourcesMap)
}

// Source returns the golang-migrate source URL registered under name, or an
// empty string if none is configured for that name.
func (c Config) Source(name string) string {
	if value, ok := c.SourcesMap[name]; ok {
		return value
	}

	return ""
}

// Databases returns the configured database names, in no particular order.
func (c Config) Databases() []string {
	return lo.Keys(c.DatabasesMap)
}

// Database returns the golang-migrate database URL registered under name, or
// an empty string if none is configured for that name.
func (c Config) Database(name string) string {
	if value, ok := c.DatabasesMap[name]; ok {
		return value
	}

	return ""
}
