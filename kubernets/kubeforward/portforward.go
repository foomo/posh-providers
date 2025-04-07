package kubeforward

type PortForward struct {
	// Target cluster
	Cluster string `json:"cluster" yaml:"cluster"`
	// Target namespace
	Namespace string `json:"namespace" yaml:"namespace"`
	// Optional description
	Description string `json:"description" yaml:"description"`
	// Target name
	Target string `json:"target" yaml:"target"`
	// Target and host port mapping
	Port string `json:"port" yaml:"port"`
}
