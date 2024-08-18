package beam

import (
	"github.com/foomo/posh-providers/onepassword"
)

type Cluster struct {
	Port       int                `yaml:"port"`
	Hostname   string             `yaml:"hostname"`
	Kubeconfig onepassword.Secret `yaml:"kubeconfig"`
}
