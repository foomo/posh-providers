package licensefinder

type Config struct {
	LogPath       string   `json:"logPath" yaml:"logPath"`
	DecisionsPath string   `json:"decisionsPath" yaml:"decisionsPath"`
	Sources       []string `json:"sources" yaml:"sources"`
}
