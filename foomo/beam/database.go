package beam

type Database struct {
	// Local port the tunnel binds on 127.0.0.1
	Port int `json:"port" yaml:"port"`
	// Cloudflare Access hostname the tunnel connects to
	Hostname string `json:"hostname" yaml:"hostname"`
}
