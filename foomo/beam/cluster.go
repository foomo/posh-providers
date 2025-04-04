package beam

import (
	"github.com/foomo/posh-providers/onepassword"
)

type Cluster struct {
	Port       int                `json:"port" yaml:"port"`
	Hostname   string             `json:"hostname" yaml:"hostname"`
	Kubeconfig onepassword.Secret `json:"kubeconfig" yaml:"kubeconfig"`
}
