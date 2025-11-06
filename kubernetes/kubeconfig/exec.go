package kubeconfig

type Exec struct {
	APIVersion string   `yaml:"apiVersion"`
	Command    string   `yaml:"command"`
	Env        any      `yaml:"env"`
	Args       []string `yaml:"args"`
}
