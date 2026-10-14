package az

// Artifactory identifies an Azure Container Registry by name and resource
// group.
type Artifactory struct {
	// Name of the artifactory
	Name string `json:"name" yaml:"name"`
	// ResourceGroup of the artifactory
	ResourceGroup string `json:"resourceGroup" yaml:"resourceGroup"`
}
