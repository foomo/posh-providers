package kubectl

// Config holds the kubectl provider's settings.
type Config struct {
	// ConfigPath to store kubeconfigs
	ConfigPath string `json:"configPath" yaml:"configPath"`
}
