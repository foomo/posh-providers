package az

type Cluster struct {
	// Name of the cluster
	Name string `json:"name" yaml:"name"`
	// Resource group of the cluster
	ResourceGroup string `json:"resourceGroup" yaml:"resourceGroup"`
	// ProxyURL to use for the kubeconfig
	ProxyURL string `json:"proxyUrl" yaml:"proxyUrl"`
}
