package squadron

type Config struct {
	// Path to the squadron root
	Path string `json:"path" yaml:"path"`
	// Cluster configurations
	Clusters []Cluster `json:"clusters" yaml:"clusters"`
}

func (c Config) Cluster(name string) (Cluster, bool) {
	for _, cluster := range c.Clusters {
		if cluster.Name == name {
			return cluster, true
		}
	}
	return Cluster{}, false
}

func (c Config) ClusterNames() []string {
	var ret []string
	for _, cluster := range c.Clusters {
		ret = append(ret, cluster.Name)
	}
	return ret
}
