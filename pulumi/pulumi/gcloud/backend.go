package pulumi

import (
	"github.com/foomo/posh-providers/onepassword"
)

type Backend struct {
	// Location is the Google Cloud location the bucket is created in. Only read
	// by `backend create`; it is not used when logging in or running a stack.
	Location string `json:"location" yaml:"location"`
	// Bucket is the Google Cloud Storage bucket holding the pulumi state. Used
	// as gs://<bucket> for both `backend login` and PULUMI_BACKEND_URL.
	Bucket string `json:"bucket" yaml:"bucket"`
	// Project is the Google Cloud project owning the bucket. Exported as
	// GOOGLE_PROJECT for every stack operation.
	Project string `json:"project" yaml:"project"`
	// Passphrase resolves the pulumi config passphrase from 1Password and is
	// exported as PULUMI_CONFIG_PASSPHRASE. Every stack operation resolves it,
	// so all of them require an unlocked 1Password session.
	Passphrase onepassword.Secret `json:"passphrase" yaml:"passphrase"`
}
