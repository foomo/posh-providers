package licensefinder

type Config struct {
	// Directory license_finder writes its log to, passed as `--log-directory`.
	LogPath string `json:"logPath" yaml:"logPath"`
	// File recording the approval decisions, passed as `--decisions-file`. Every
	// `add`/`remove` verb rewrites this file, so it is the provider's only
	// persistent state and belongs in version control.
	DecisionsPath string `json:"decisionsPath" yaml:"decisionsPath"`
	// Dependency-manifest filenames to search the project for, e.g. `go.mod` or
	// `yarn.lock`. Each match's containing directory becomes a scanned project;
	// the manifests themselves are never read here.
	Sources []string `json:"sources" yaml:"sources"`
}
