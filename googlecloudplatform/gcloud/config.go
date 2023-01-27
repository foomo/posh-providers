package gcloud

type Config struct {
	Project    string    `json:"project" yaml:"project"`
	ConfigPath string    `json:"configPath" yaml:"configPath"`
	Clusters   []Cluster `json:"clusters" yaml:"clusters"`
}

func (c Config) ClusterNames() []string {
	var ret []string
	for _, cluster := range c.Clusters {
		ret = append(ret, cluster.Name)
	}
	return ret
}

func (c Config) ClustersByName(name string) (Cluster, bool) {
	for _, cluster := range c.Clusters {
		if cluster.Name == name {
			return cluster, true
		}
	}
	return Cluster{}, false
}

type Cluster struct {
	Name   string `json:"name" yaml:"name"`
	Region string `json:"region" yaml:"region"`
}
