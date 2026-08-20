package pulumi

import (
	"github.com/foomo/posh-providers/onepassword"
)

type Backend struct {
	// Location is the Azure region the resource group and storage account are
	// created in. Only read by `backend create`; it is not used when logging in
	// or running a stack.
	Location string `json:"location" yaml:"location"`
	// Container is the blob container holding the pulumi state. Used as
	// azblob://<container> for both `backend login` and PULUMI_BACKEND_URL.
	Container string `json:"container" yaml:"container"`
	// Subscription is the Azure subscription ID owning the backend. Exported as
	// ARM_SUBSCRIPTION_ID for every stack operation, so it also decides which
	// subscription the stack's own resources are created in.
	Subscription string `json:"subscription" yaml:"subscription"`
	// ResourceGroup holds the storage account. Read by `backend create` and by
	// the storage key lookup every stack operation performs.
	ResourceGroup string `json:"resourceGroup" yaml:"resourceGroup"`
	// StorageAccount holds the container. Exported as AZURE_STORAGE_ACCOUNT,
	// and its access key is fetched from Azure on every stack operation.
	StorageAccount string `json:"storageAccount" yaml:"storageAccount"`
	// Passphrase resolves the pulumi config passphrase from 1Password and is
	// exported as PULUMI_CONFIG_PASSPHRASE. Every stack operation resolves it,
	// so all of them require an unlocked 1Password session.
	Passphrase onepassword.Secret `json:"passphrase" yaml:"passphrase"`
}
