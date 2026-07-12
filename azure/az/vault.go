package az

type Vault struct {
	// Name of the key vault
	Name string `json:"name" yaml:"name"`
	// Resource group of the key vault
	ResourceGroup string `json:"resourceGroup" yaml:"resourceGroup"`
}
