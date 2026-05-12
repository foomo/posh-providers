package ssh

// PortForward represents a configuration for setting up an SSH-based port forwarding session.
type PortForward struct {
	// Local port to bind (0 = auto-assign)
	Port int `json:"port" yaml:"port"`
	// Target server proxy host
	Host string `json:"host" yaml:"host"`
	// Target server proxy port
	HostPort int `json:"hostPort" yaml:"hostPort"`
	// SSH server username
	Username string `json:"username" yaml:"username"`
	// path to Username private key (-i)
	IdentityFile string `json:"identityFile"  yaml:"identityFile"`
	// Username agent socket path (-o IdentityAgent)
	IdentityAgent string `json:"identityAgent" yaml:"identityAgent"`
}
