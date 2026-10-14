package ssh

import (
	"maps"
	"slices"
)

// Config holds the named port forwards and SOCKS5 tunnels available to the
// ssh command.
type Config struct {
	PortForwards  map[string]PortForward  `json:"portForwards" yaml:"portForwards"`
	Socks5Tunnels map[string]Socks5Tunnel `json:"socks5Tunnels" yaml:"socks5Tunnels"`
}

// PortForward returns the named port forward configuration and whether it exists.
func (c Config) PortForward(name string) (PortForward, bool) {
	t, ok := c.PortForwards[name]
	return t, ok
}

// PortForwardNames returns the configured port forward names, sorted.
func (c Config) PortForwardNames() []string {
	return slices.Sorted(maps.Keys(c.PortForwards))
}

// Socks5Tunnel returns the named SOCKS5 tunnel configuration and whether it exists.
func (c Config) Socks5Tunnel(name string) (Socks5Tunnel, bool) {
	t, ok := c.Socks5Tunnels[name]
	return t, ok
}

// Socks5TunnelNames returns the configured SOCKS5 tunnel names, sorted.
func (c Config) Socks5TunnelNames() []string {
	return slices.Sorted(maps.Keys(c.Socks5Tunnels))
}
