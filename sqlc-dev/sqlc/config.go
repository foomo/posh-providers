package sqlc

type Config struct {
	TempDir  string `json:"tempDir" yaml:"tempDir"`
	CacheDir string `json:"cacheDir" yaml:"cacheDir"`
}
