package mkcert

type Config struct {
	CertificatePath string        `yaml:"certificatePath"`
	Certificates    []Certificate `yaml:"certificates"`
}
