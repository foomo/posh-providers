package cloudflared

// Access configures one `cloudflared access` tunnel to a Cloudflare Access
// target.
type Access struct {
	// Access mode passed to `cloudflared access`, e.g. tcp or ssh
	Type string `json:"type" yaml:"type"`
	// Hostname of the Cloudflare Access target to tunnel to
	Hostname string `json:"hostname" yaml:"hostname"`
	// Local port the tunnel listens on, bound to 127.0.0.1
	Port int `json:"port" yaml:"port"`
}
