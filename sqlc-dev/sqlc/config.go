package sqlc

// Config holds the sqlc provider's configuration.
type Config struct {
	// Directory passed to sqlc as SQLCTMPDIR, resolved relative to the project root
	TempDir string `json:"tempDir" yaml:"tempDir"`
	// Directory passed to sqlc as SQLCCACHE, resolved relative to the project root
	CacheDir string `json:"cacheDir" yaml:"cacheDir"`
}
