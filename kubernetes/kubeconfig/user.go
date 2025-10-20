package kubeconfig

type User struct {
	Name string   `yaml:"name"`
	User UserData `yaml:"user"`
}
