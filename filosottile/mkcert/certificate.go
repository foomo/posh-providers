package mkcert

type Certificate struct {
	Name  string   `json:"name" yaml:"name"`
	Names []string `json:"names" yaml:"names"`
}
