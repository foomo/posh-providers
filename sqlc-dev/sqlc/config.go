package sqlc

type Config struct {
	TempDir  string `json:"tempDir" yaml:"tempDir"`
	CacheDir string `json:"cacheDir" yaml:"cacheDir"`
	// FindIgnore contains regular expressions for file and directory names to skip when discovering sqlc.yaml files.
	FindIgnore []string `json:"findIgnore" yaml:"findIgnore"`
}
