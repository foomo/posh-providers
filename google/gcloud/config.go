package gcloud

type Config struct {
	ConfigDir    string        `json:"configDir" yaml:"configDir"`
	Environments []Environment `json:"environments" yaml:"environments"`
}

type Environment struct {
	Name     string    `json:"name" yaml:"name"`
	Clusters []Cluster `json:"clusters" yaml:"clusters"`
}

func (c Config) ClusterNames() []string {
	var ret []string
	for _, account := range c.Environments {
		for _, cluster := range account.Clusters {
			ret = append(ret, cluster.Name)
		}
	}
	return ret
}

type Cluster struct {
	Project string `json:"project" yaml:"project"`
	Region  string `json:"region" yaml:"region"`
	Name    string `json:"name" yaml:"name"`
}
