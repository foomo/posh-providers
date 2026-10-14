package k3d

type (
	// Cluster is a running k3d cluster as reported by `k3d cluster list`.
	Cluster struct {
		Name string `json:"name"`
	}
)
