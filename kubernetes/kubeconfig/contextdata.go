package kubeconfig

type ContextData struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}
