package stackit

type Cluster struct {
	// Real SKE cluster name, passed to the CLI as the positional argument. The
	// kubeconfig filename comes from the map key instead, unless
	// `CommandWithClusterNameFn` overrides it.
	Name string `json:"name"`
}
