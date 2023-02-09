package gcloud

type Config struct {
	ConfigDir    string        `json:"configDir" yaml:"configDir"`
	Environments []Environment `json:"environments" yaml:"environments"`
}

func (c Config) FindCluster(envName, clusterName string) (Cluster, bool) {
	for _, env := range c.Environments {
		if env.Name != envName {
			continue
		}

		for _, cluster := range env.Clusters {
			if cluster.Name == clusterName {
				return cluster, true
			}
		}
	}
	return Cluster{}, false
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

func (c Config) ClusterNamesForEnv(envName string) []string {
	for _, env := range c.Environments {
		if env.Name == envName {
			names := make([]string, len(env.Clusters))
			for idx, cluster := range env.Clusters {
				names[idx] = cluster.Name
			}
			return names
		}
	}
	return nil
}

type Cluster struct {
	Project string `json:"project" yaml:"project"`
	Region  string `json:"region" yaml:"region"`
	Name    string `json:"name" yaml:"name"`
}
