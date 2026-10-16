package gnat

import (
	"github.com/foomo/posh-providers/kubernetes/kubectl"
)

// Profile identifies a NATS server by URL, with a description shown in
// shell autocompletion.
type Profile struct {
	// URL is the NATS server URL, passed to the CLI via -url.
	URL string `json:"url,omitempty,omitempty" yaml:"url,omitempty"`
	// URLForward replaces URL with the local address of a kubectl port-forward
	// to the given target. The profile name is used as the kubectl cluster name.
	URLForward *kubectl.PortForwardTarget `json:"urlForward,omitempty" yaml:"urlForward,omitempty"`
	// Description is shown in shell autocompletion.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Kubeforward is the port forwarding configuration for the NATS server.
	Kubeforward *kubectl.PortForwardTarget `json:"kubeforward,omitempty" yaml:"kubeforward,omitempty"`
}
