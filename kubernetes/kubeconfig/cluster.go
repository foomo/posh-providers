package kubeconfig

type Cluster struct {
	Name    string      `yaml:"name"`
	Cluster ClusterData `yaml:"cluster"`
}
