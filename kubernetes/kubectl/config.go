package kubectl

type ClusterConfig struct {
	Proxy string `json:"proxy" yaml:"proxy" jsonschema:"description=Named SSH proxy from the top-level proxies map"`
}

type Config struct {
	ConfigPath string                   `json:"configPath" yaml:"configPath"`
	Clusters   map[string]ClusterConfig `json:"clusters"   yaml:"clusters"`
}

func (c Config) ClusterProxy(name string) string {
	if cc, ok := c.Clusters[name]; ok {
		return cc.Proxy
	}

	return ""
}
