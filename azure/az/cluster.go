package az

// Cluster identifies an AKS cluster by name, resource group, and an optional
// proxy URL to write into its retrieved kubeconfig.
type Cluster struct {
	// Name of the cluster
	Name string `json:"name" yaml:"name"`
	// Resource group of the cluster
	ResourceGroup string `json:"resourceGroup" yaml:"resourceGroup"`
	// ProxyURL to use for the kubeconfig
	ProxyURL string `json:"proxyUrl" yaml:"proxyUrl"`
}
