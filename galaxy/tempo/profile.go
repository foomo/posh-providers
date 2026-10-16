package tempo

import (
	"github.com/foomo/posh-providers/kubernetes/kubectl"
)

// Profile identifies a Temporal server by address, with a description shown
// in shell autocompletion.
type Profile struct {
	// URL is the Temporal server address (host:port), passed to the CLI via --address.
	URL string `json:"url,omitempty,omitempty" yaml:"url,omitempty"`
	// URLForward dynamically uses kubernetes port forwarding
	URLForward *kubectl.PortForwardTarget `json:"urlForward,omitempty" yaml:"urlForward,omitempty"`
	// Description is shown in shell autocompletion.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Kubeforward is the port forwarding configuration for the NATS server.
	Kubeforward *kubectl.PortForwardTarget `json:"kubeforward,omitempty" yaml:"kubeforward,omitempty"`
}
