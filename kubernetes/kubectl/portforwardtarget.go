package kubectl

// PortForwardTarget identifies the pod or service a kubectl port-forward is
// opened against to reach a server only reachable from inside the cluster.
type PortForwardTarget struct {
	// Cluster to use for the kubectl command, optional; defaults to the
	// consumer's profile or env name
	Cluster string `json:"cluster" yaml:"cluster"`
	// Port on the target
	Port int `json:"port" yaml:"port"`
	// Resource pod or service, e.g. svc/grafana
	Resource string `json:"resource" yaml:"resource"`
	// Namespace of the target pod or service
	Namespace string `json:"namespace" yaml:"namespace"`
	// Profile to use for the kubectl command, optional
	Profile string `json:"profile" yaml:"profile"`
}
