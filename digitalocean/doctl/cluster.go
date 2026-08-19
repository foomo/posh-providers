package doctl

type Cluster struct {
	// Cluster ID or name passed to doctl. The README's sample uses the cluster
	// UUID, which is also what the generated kubeconfig's credential plugin
	// requires. The kubeconfig filename comes from the map key instead, unless
	// `CommandWithClusterNameFn` overrides it.
	Name string `json:"name" yaml:"name"`
}
