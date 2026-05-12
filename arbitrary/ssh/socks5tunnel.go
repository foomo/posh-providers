package ssh

// Socks5Tunnel represents a configuration for an SSH-based proxy with optional HTTP support.
type Socks5Tunnel struct {
	// Local port to bind (0 = auto-assign)
	Port int `json:"port" yaml:"port"`
	// Target server proxy host
	Host string `json:"host" yaml:"host"`
	// Target server proxy port
	HostPort int `json:"hostPort" yaml:"hostPort"`
	// SSH server username
	Username string `json:"username" yaml:"username"`
	// path to Username private key (-i)
	IdentityFile string `json:"identityFile" yaml:"identityFile"`
	// Username agent socket path (-o IdentityAgent)
	IdentityAgent string `json:"identityAgent" yaml:"identityAgent"`
}
