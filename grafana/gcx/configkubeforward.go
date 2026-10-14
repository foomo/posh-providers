package gcx

// ConfigKubeForward identifies the pod or service a kubectl port-forward is
// opened against to reach a Grafana server only reachable from inside the
// cluster.
type ConfigKubeForward struct {
	// Port on the target
	Port int `json:"port" yaml:"port"`
	// Target pod or service
	Target string `json:"target" yaml:"target"`
	// Namespace of the target pod or service
	Namespace string `json:"namespace" yaml:"namespace"`
}
