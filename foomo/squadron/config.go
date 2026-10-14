package squadron

// Config holds the squadron provider's configuration.
type Config struct {
	// Path to the squadron root
	Path string `json:"path" yaml:"path"`
	// Cluster configurations
	Clusters []Cluster `json:"clusters" yaml:"clusters"`
}

// Cluster returns the Cluster configured under name and reports whether it
// was found.
func (c Config) Cluster(name string) (Cluster, bool) {
	for _, cluster := range c.Clusters {
		if cluster.Name == name {
			return cluster, true
		}
	}

	return Cluster{}, false
}

// ClusterNames returns the configured cluster names, in configuration order.
func (c Config) ClusterNames() []string {
	var ret []string
	for _, cluster := range c.Clusters {
		ret = append(ret, cluster.Name)
	}

	return ret
}
