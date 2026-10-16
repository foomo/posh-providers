package gcx

import (
	"github.com/foomo/posh-providers/kubernetes/kubectl"
)

// ConfigEnv is a single gcx environment: the Grafana server, org, credential
// source, optional urlForward target, and extra environment variables
// passed to the gcx CLI.
type ConfigEnv struct {
	// URL of the grafana server
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
	// URLForward replaces URL with the local address of a kubectl port-forward
	// to the given target. Its cluster, or the env name if unset, is used as the
	// kubectl cluster name.
	URLForward *kubectl.PortForwardTarget `json:"urlForward,omitempty" yaml:"urlForward,omitempty"`
	// Env to pass to the cli
	Env []string `json:"env" yaml:"env"`
	// OrgID to set
	OrgID string `json:"orgId" yaml:"orgId"`
	// Token to set
	Token string `json:"token" yaml:"token"`
	// TokenCmd to execute returning the token
	TokenCmd []string `json:"tokenCmd" yaml:"tokenCmd"`
}
