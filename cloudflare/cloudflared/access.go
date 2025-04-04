package cloudflared

type Access struct {
	Type     string `json:"type" yaml:"type"`
	Hostname string `json:"hostname" yaml:"hostname"`
	Port     int    `json:"port" yaml:"port"`
}
