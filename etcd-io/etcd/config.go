package etcd

type Config struct {
	// Directory the `edit` verb writes its scratch copy of a value into, before and
	// after editing. Created at shell startup; resolved against the project root, so
	// the file — which holds real etcd contents — lands inside the checkout.
	ConfigPath string `json:"configPath" yaml:"configPath"`
	// Clusters this command can reach. Only clusters listed here are offered, and
	// completion additionally requires a matching kubeconfig to exist.
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

type Cluster struct {
	// Cluster name, matched against both the `cluster` argument and the kubectl
	// cluster whose kubeconfig is used to reach the pod.
	Name string `json:"name" yaml:"name"`
	// Name of the etcd pod to exec into. Must be an exact pod name, not a
	// deployment or label selector.
	PodName string `json:"podName" yaml:"podName"`
	// Namespace the etcd pod runs in.
	Namespace string `json:"namespace" yaml:"namespace"`
	// etcd keys offered as completions for the `path` argument. Advisory only — any
	// key can be typed, whether listed or not.
	Paths []string `json:"paths" yaml:"paths"`
}
