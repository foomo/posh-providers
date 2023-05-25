package kubeconfig

type Exec struct {
	APIVersion string      `yaml:"apiVersion"`
	Command    string      `yaml:"command"`
	Env        interface{} `yaml:"env"`
	Args       []string    `yaml:"args"`
}
