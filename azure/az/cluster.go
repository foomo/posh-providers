package az

type Cluster struct {
	// Cluster
	Name string `json:"name" yaml:"name"`
	// Cluster resource group name
	ResourceGroup string `json:"resourceGroup" yaml:"resourceGroup"`
	// ProxyURL
	ProxyURL string `json:"proxyUrl" yaml:"proxyUrl"`
}
