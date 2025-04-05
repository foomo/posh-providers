package mkcert

type Config struct {
	CertificatePath string        `json:"certificatePath" yaml:"certificatePath"`
	Certificates    []Certificate `json:"certificates" yaml:"certificates"`
}
