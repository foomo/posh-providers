package k3d

type (
	// Registry is a k3d container registry as reported by `k3d registry list`.
	Registry struct {
		Name string `json:"name"`
	}
)
