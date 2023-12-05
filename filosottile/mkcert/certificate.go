package mkcert

type Certificate struct {
	Name  string   `yaml:"name"`
	Names []string `yaml:"names"`
}
