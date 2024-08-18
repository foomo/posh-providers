package cloudflared

type Access struct {
	Type     string `yaml:"type"`
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
}
