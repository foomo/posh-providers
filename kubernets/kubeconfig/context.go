package kubeconfig

type Context struct {
	Name    string      `yaml:"name"`
	Context ContextData `yaml:"context"`
}
