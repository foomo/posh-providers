package kubectl

type Config struct {
	// ConfigPath to store kubeconfigs
	ConfigPath string `json:"configPath" yaml:"configPath"`
}
