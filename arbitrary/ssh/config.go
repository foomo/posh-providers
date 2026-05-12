package ssh

import (
	"maps"
	"slices"
)

type Config struct {
	PortForwards  map[string]PortForward  `json:"portForwards" yaml:"portForwards"`
	Socks5Tunnels map[string]Socks5Tunnel `json:"socks5Tunnels" yaml:"socks5Tunnels"`
}

func (c Config) PortForward(name string) (PortForward, bool) {
	t, ok := c.PortForwards[name]
	return t, ok
}

func (c Config) PortForwardNames() []string {
	return slices.Sorted(maps.Keys(c.PortForwards))
}

func (c Config) Socks5Tunnel(name string) (Socks5Tunnel, bool) {
	t, ok := c.Socks5Tunnels[name]
	return t, ok
}

func (c Config) Socks5TunnelNames() []string {
	return slices.Sorted(maps.Keys(c.Socks5Tunnels))
}
