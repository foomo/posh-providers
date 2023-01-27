package etcd

type Config struct {
	ConfigPath string    `json:"configPath" yaml:"configPath"`
	Clusters   []Cluster `json:"clusters" yaml:"clusters"`
}

func (c Config) Cluster(name string) (Cluster, bool) {
	for _, cluster := range c.Clusters {
		if cluster.Name == name {
			return cluster, true
		}
	}
	return Cluster{}, false
}

type Cluster struct {
	Name      string   `json:"name" yaml:"name"`
	PodName   string   `json:"podName" yaml:"podName"`
	Namespace string   `json:"namespace" yaml:"namespace"`
	Paths     []string `json:"paths" yaml:"paths"`
}
