package sshtunnel

type (
	Config struct {
		SocketsDir string   `json:"socketsDir" yaml:"socketsDir"`
		TempDir    string   `json:"tempDir" yaml:"tempDir"`
		Tunnels    []Tunnel `json:"tunnels" yaml:"tunnels"`
	}
)

// Tunnel returns a tunnel by name
func (c Config) Tunnel(name string) (Tunnel, bool) {
	for _, t := range c.Tunnels {
		if t.Name == name {
			return t, true
		}
	}
	return Tunnel{}, false
}

// TunnelNames returns all configured tunnel names
func (c Config) TunnelNames() []string {
	var ret []string
	for _, tunnel := range c.Tunnels {
		ret = append(ret, tunnel.Name)
	}
	return ret
}
