package kubeconfig

type UserData struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
	Exec                  Exec   `yaml:"exec"`
}
