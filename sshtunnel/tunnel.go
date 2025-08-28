package sshtunnel

type (
	Tunnel struct {
		// Unique tunnel name
		Name string `json:"name" yaml:"name"`
		// Whether this tunnel requires sudo privileges
		Sudo bool `json:"sudo,omitempty" yaml:"sudo,omitempty"`
		// Local port to bind
		LocalPort int `json:"localPort" yaml:"localPort"`
		// Target server proxy host
		TargetProxyHost string `json:"targetProxyHost" yaml:"targetProxyHost"`
		// Target server proxy port
		TargetProxyPort int `json:"targetProxyPort" yaml:"targetProxyPort"`
		// SSH server username
		TargetUsername string `json:"targetUsername" yaml:"targetUsername"`
		// SSH server hostname or IP
		TargetHost string `json:"targetHost" yaml:"targetHost"`
		// Authentication details (password, private key)
		TargetAuth TargetAuth `json:"targetAuth,omitzero" yaml:"targetAuth,omitempty"`
	}

	// TargetAuth holds authentication information for SSH Target Server
	TargetAuth struct {
		// Auth method: "sshpass", "key",
		Type string `json:"type,omitempty" yaml:"type,omitempty"`
		// Password-based authentication (optional)
		Password string `json:"password,omitempty" yaml:"password,omitempty"`
		// Private key path (optional)
		PrivateKey string `json:"privateKey,omitempty" yaml:"privateKey,omitempty"`
	}
)
