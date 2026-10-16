package kubectl

// PortForwardTarget identifies the pod or service a kubectl port-forward is
// opened against to reach a Grafana server only reachable from inside the
// cluster.
type PortForwardTarget struct {
	// Port on the target
	Port int `json:"port" yaml:"port"`
	// Resource pod or service
	Resource string `json:"resource" yaml:"resource"`
	// Namespace of the target pod or service
	Namespace string `json:"namespace" yaml:"namespace"`
	// Profile to use for the kubectl command
	Profile string `json:"profile" yaml:"profile"`
}
