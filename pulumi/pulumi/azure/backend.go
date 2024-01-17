package pulumi

type Backend struct {
	Location       string `json:"location" yaml:"location"`
	Container      string `json:"container" yaml:"container"`
	Subscription   string `json:"subscription" yaml:"subscription"`
	ResourceGroup  string `json:"resourceGroup" yaml:"resourceGroup"`
	StorageAccount string `json:"storageAccount" yaml:"storageAccount"`
}
