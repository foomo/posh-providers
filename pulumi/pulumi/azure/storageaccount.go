package pulumi

// StorageAccount identifies an Azure storage account by name.
type StorageAccount struct {
	// Name is the storage account's name.
	Name string `json:"name" yaml:"name"`
}
