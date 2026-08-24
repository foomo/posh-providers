package beam

import (
	"github.com/foomo/posh-providers/onepassword"
)

type Cluster struct {
	// Local port the tunnel binds on 127.0.0.1
	Port int `json:"port" yaml:"port"`
	// Cloudflare Access hostname the tunnel connects to
	Hostname string `json:"hostname" yaml:"hostname"`
	// 1Password document holding the kubeconfig; $PORT in it is replaced with the port above
	Kubeconfig onepassword.Secret `json:"kubeconfig" yaml:"kubeconfig"`
}
