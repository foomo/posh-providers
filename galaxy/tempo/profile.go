package tempo

import (
	"github.com/foomo/posh-providers/kubernetes/kubectl"
)

// Profile identifies a Temporal server by address, with a description shown
// in shell autocompletion.
type Profile struct {
	// URL is the Temporal server address (host:port), passed to the CLI via --address.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
	// URLForward replaces URL with the local address of a kubectl port-forward
	// to the given target. Its cluster, or the profile name if unset, is used as
	// the kubectl cluster name.
	URLForward *kubectl.PortForwardTarget `json:"urlForward,omitempty" yaml:"urlForward,omitempty"`
	// Description is shown in shell autocompletion.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Kubeforward is the port forwarding configuration for the Temporal server.
	Kubeforward *kubectl.PortForwardTarget `json:"kubeforward,omitempty" yaml:"kubeforward,omitempty"`
}
